import { useCallback, useState } from "react";

type ExpandedNodes = {
	collapse: (assetId: string) => void;
	expand: (assetId: string) => void;
	expandedAssetIds: ReadonlySet<string>;
	toggle: (assetId: string) => void;
};

type Parameters = {
	selectedAssetId: null | string;
	/** The IDs of the selected asset's path, substation first; `null` until loaded. */
	selectedPath: null | readonly string[];
};

const addAssetIds = (
	assetIds: ReadonlySet<string>,
	added: readonly string[],
): ReadonlySet<string> => new Set([...assetIds, ...added]);

const removeAssetId = (
	assetIds: ReadonlySet<string>,
	assetId: string,
): ReadonlySet<string> => {
	const next = new Set(assetIds);
	next.delete(assetId);

	return next;
};

/**
 * The tree's expanded nodes. When the selected asset's path is known, its
 * ancestors are expanded once per selection, so the asset is visible (spec
 * FR-005, FR-019) and the user can still collapse them afterwards. This
 * adjusts state during render when the selection changes (research R8); no
 * effect copies query data into state.
 */
const useExpandedNodes = ({
	selectedAssetId,
	selectedPath,
}: Parameters): ExpandedNodes => {
	const [expandedAssetIds, setExpandedAssetIds] = useState<ReadonlySet<string>>(
		() => new Set(),
	);
	const [revealedAssetId, setRevealedAssetId] = useState<null | string>(null);

	const isPathOfSelection =
		selectedAssetId !== null &&
		selectedPath !== null &&
		selectedPath.includes(selectedAssetId);

	if (isPathOfSelection && revealedAssetId !== selectedAssetId) {
		setRevealedAssetId(selectedAssetId);
		setExpandedAssetIds((current) =>
			addAssetIds(
				current,
				selectedPath.filter((assetId) => assetId !== selectedAssetId),
			),
		);
	}

	const expand = useCallback((assetId: string): void => {
		setExpandedAssetIds((current) => addAssetIds(current, [assetId]));
	}, []);

	const collapse = useCallback((assetId: string): void => {
		setExpandedAssetIds((current) => removeAssetId(current, assetId));
	}, []);

	const toggle = useCallback((assetId: string): void => {
		setExpandedAssetIds((current) =>
			current.has(assetId)
				? removeAssetId(current, assetId)
				: addAssetIds(current, [assetId]),
		);
	}, []);

	return { collapse, expand, expandedAssetIds, toggle };
};

export { useExpandedNodes };
