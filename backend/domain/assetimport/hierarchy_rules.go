package assetimport

import (
	"sort"
	"strings"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/valueobject"
)

// fileIndex is what the hierarchy rules know about the whole file. Every
// decision uses it, never the row sequence, so the result is the same for
// any row order (FR-017).
type fileIndex struct {
	idCount    map[string]int
	uniqueByID map[string]*candidate
	cycles     map[string][]string
}

// checkHierarchy applies the parent rules (spec FR-011 – FR-015, FR-016a) to
// every row whose own fields are valid, and lists every rule a row breaks:
//
//  1. SUBSTATION_HAS_PARENT / PARENT_REQUIRED / SELF_PARENT /
//     PARENT_NOT_FOUND — the parent reference itself;
//  2. INVALID_PARENT_TYPE — against the in-file row or the stored asset;
//  3. CYCLE_DETECTED — every row on a parent-pointer cycle.
//
// PARENT_REJECTED (FR-016) comes later, from propagateRejectedParents, once
// the re-import rules have also had their say.
func checkHierarchy(candidates []*candidate, stored StoredAssets) fileIndex {
	idCount := map[string]int{}
	for _, c := range candidates {
		if c.assetID != "" {
			idCount[c.assetID]++
		}
	}

	// uniqueByID holds the rows whose asset ID occurs exactly once; a parent
	// ID shared by several rows names no single parent.
	uniqueByID := map[string]*candidate{}
	for _, c := range candidates {
		if idCount[c.assetID] == 1 {
			uniqueByID[c.assetID] = c
		}
	}

	cycles := findCycles(uniqueByID)

	for _, c := range candidates {
		if c.fieldsValid() {
			checkParentReference(c, idCount, uniqueByID, stored, cycles)
		}
	}

	return fileIndex{idCount: idCount, uniqueByID: uniqueByID, cycles: cycles}
}

func checkParentReference(
	c *candidate,
	idCount map[string]int,
	uniqueByID map[string]*candidate,
	stored StoredAssets,
	cycles map[string][]string,
) {
	assetType := c.asset.AssetType
	parentID := c.parentAssetID

	if assetType == valueobject.AssetTypeSubstation {
		if parentID != "" {
			c.hierarchyIssues = append(c.hierarchyIssues, Issue{Code: errorcode.SubstationHasParent, Args: []any{parentID}})
		}
		return
	}
	if parentID == "" {
		c.hierarchyIssues = append(c.hierarchyIssues, Issue{Code: errorcode.ParentRequired, Args: []any{assetType.String()}})
		return
	}
	if parentID == c.assetID {
		c.hierarchyIssues = append(c.hierarchyIssues, Issue{Code: errorcode.SelfParent})
		return
	}

	storedParent := stored.ByAssetID[parentID]
	if idCount[parentID] == 0 && storedParent == nil {
		c.hierarchyIssues = append(c.hierarchyIssues, Issue{Code: errorcode.ParentNotFound, Args: []any{parentID}})
		return
	}

	if parentType, known := parentTypeOf(parentID, idCount, uniqueByID, storedParent); known {
		if permitted, _ := assetType.PermittedParent(); parentType != permitted {
			c.hierarchyIssues = append(c.hierarchyIssues, Issue{
				Code: errorcode.InvalidParentType,
				Args: []any{assetType.String(), permitted.String(), parentID, parentType.String()},
			})
		}
	}

	if members, inCycle := cycles[c.assetID]; inCycle {
		c.hierarchyIssues = append(c.hierarchyIssues, Issue{
			Code: errorcode.CycleDetected,
			Args: []any{c.assetID, strings.Join(members, ", ")},
		})
	}
}

// parentTypeOf returns the parent's asset type when it is a single in-file
// row with a supported type, or a stored asset. A duplicated or unsupported
// in-file parent has no usable type; its rejection reaches the child as
// PARENT_REJECTED instead.
func parentTypeOf(
	parentID string,
	idCount map[string]int,
	uniqueByID map[string]*candidate,
	storedParent *entity.Asset,
) (valueobject.AssetType, bool) {
	if idCount[parentID] > 1 {
		return "", false
	}
	if parent, inFile := uniqueByID[parentID]; inFile {
		parentType, err := valueobject.ParseAssetType(parent.assetType)
		return parentType, err == nil
	}
	if storedParent != nil {
		return storedParent.AssetType, true
	}
	return "", false
}

// findCycles follows parent pointers between rows with a unique asset ID and
// returns, for every row on a cycle, the sorted IDs of its cycle. A row that
// names itself is SELF_PARENT, not a cycle.
func findCycles(uniqueByID map[string]*candidate) map[string][]string {
	const (
		unvisited = iota
		onPath
		done
	)

	state := map[string]int{}
	cycles := map[string][]string{}

	ids := make([]string, 0, len(uniqueByID))
	for id := range uniqueByID {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, start := range ids {
		var path []string
		current := start
		selfLoop := false
		for {
			node, ok := uniqueByID[current]
			if !ok || state[current] != unvisited {
				break
			}
			state[current] = onPath
			path = append(path, current)
			if node.parentAssetID == current {
				selfLoop = true
				break
			}
			current = node.parentAssetID
		}

		if !selfLoop && state[current] == onPath {
			index := indexOf(path, current)
			members := append([]string(nil), path[index:]...)
			sort.Strings(members)
			for _, member := range members {
				cycles[member] = members
			}
		}
		for _, id := range path {
			state[id] = done
		}
	}
	return cycles
}

// propagateRejectedParents adds PARENT_REJECTED to valid-field rows whose
// in-file parent is rejected, until nothing changes, so it reaches the whole
// subtree below a rejected row. Cycle members are skipped: their parent is
// already reported as part of the cycle.
func propagateRejectedParents(candidates []*candidate, index fileIndex) {
	idCount, uniqueByID, cycles := index.idCount, index.uniqueByID, index.cycles
	for changed := true; changed; {
		changed = false
		for _, c := range candidates {
			if !c.fieldsValid() || hasIssue(c.hierarchyIssues, errorcode.ParentRejected) {
				continue
			}
			parentID := c.parentAssetID
			if parentID == "" || parentID == c.assetID || idCount[parentID] == 0 {
				continue
			}
			if _, inCycle := cycles[c.assetID]; inCycle {
				continue
			}

			parent, unique := uniqueByID[parentID]
			if !unique || parent.rejected() {
				c.hierarchyIssues = append(c.hierarchyIssues, Issue{Code: errorcode.ParentRejected, Args: []any{parentID}})
				changed = true
			}
		}
	}
}

func hasIssue(issues []Issue, code string) bool {
	for _, issue := range issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}

func indexOf(values []string, value string) int {
	for index, candidate := range values {
		if candidate == value {
			return index
		}
	}
	return -1
}
