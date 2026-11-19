package asset_usecase

import (
	"bytes"
	"io"
	"testing"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/valueobject"
	mocksrepo "smart-grid-asset-registry/internal/mocks/repository"

	"github.com/sirupsen/logrus"
)

// deps bundles the mocks one test case needs; created per case so an
// unexpected repository call fails that case.
type deps struct {
	assetRepo *mocksrepo.AssetRepository
}

func newDeps(t *testing.T) deps {
	return deps{assetRepo: mocksrepo.NewAssetRepository(t)}
}

func (d deps) usecase() assetUsecase {
	log := logrus.New()
	log.SetOutput(io.Discard)
	return assetUsecase{assetRepository: d.assetRepo, log: log}
}

// loggingUsecase captures the usecase's JSON log lines.
func (d deps) loggingUsecase() (assetUsecase, *bytes.Buffer) {
	var logs bytes.Buffer
	u := d.usecase()
	u.log = logrus.New()
	u.log.SetOutput(&logs)
	u.log.SetFormatter(&logrus.JSONFormatter{})
	return u, &logs
}

func pointer[T any](value T) *T { return &value }

func node(assetID string, assetType valueobject.AssetType, childCount int) *entity.AssetNode {
	return &entity.AssetNode{AssetID: assetID, AssetName: "Name " + assetID, AssetType: assetType, ChildCount: childCount}
}

func nodeIDs(nodes []*entity.AssetNode) []string {
	ids := make([]string, 0, len(nodes))
	for _, n := range nodes {
		ids = append(ids, n.AssetID)
	}
	return ids
}
