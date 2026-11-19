import { useCallback, useRef, useState } from "react";

import { Button } from "~/libs/components/button/button.js";
import { ButtonVariant, ControlSize } from "~/libs/enums/enums.js";
import { useGetAssetNodesQuery } from "~/modules/assets/assets.js";

import { TREE_ROOT_LEVEL } from "../../libs/constants/constants.js";
import { AssetsPageMessage } from "../../libs/enums/enums.js";
import { useTreeKeyboard } from "../../libs/hooks/use-tree-keyboard/use-tree-keyboard.hook.js";
import { TreeNode } from "./components/tree-node/tree-node.js";
import styles from "./styles.module.css";

type Properties = {
	selectedAssetId: null | string;
};

const addAssetId = (
	assetIds: ReadonlySet<string>,
	assetId: string,
): ReadonlySet<string> => new Set(assetIds).add(assetId);

const removeAssetId = (
	assetIds: ReadonlySet<string>,
	assetId: string,
): ReadonlySet<string> => {
	const next = new Set(assetIds);
	next.delete(assetId);

	return next;
};

/**
 * The asset hierarchy as a WAI-ARIA navigation tree (research R8): every
 * substation at the top level, children loaded when a node is expanded.
 * Which nodes are expanded is local state; it is not kept across refresh.
 */
const AssetTree: React.FC<Properties> = ({ selectedAssetId }: Properties) => {
	const { data, isError, isLoading, refetch } = useGetAssetNodesQuery(null);
	const treeReference = useRef<HTMLUListElement>(null);
	const [expandedAssetIds, setExpandedAssetIds] = useState<ReadonlySet<string>>(
		() => new Set(),
	);

	const handleExpand = useCallback((assetId: string): void => {
		setExpandedAssetIds((current) => addAssetId(current, assetId));
	}, []);

	const handleCollapse = useCallback((assetId: string): void => {
		setExpandedAssetIds((current) => removeAssetId(current, assetId));
	}, []);

	const handleToggle = useCallback((assetId: string): void => {
		setExpandedAssetIds((current) =>
			current.has(assetId)
				? removeAssetId(current, assetId)
				: addAssetId(current, assetId),
		);
	}, []);

	const { focusedAssetId, handleFocus, handleKeyDown } = useTreeKeyboard({
		onCollapse: handleCollapse,
		onExpand: handleExpand,
		preferredAssetId: selectedAssetId,
		treeReference,
	});

	const handleRetryClick = useCallback((): void => {
		void refetch();
	}, [refetch]);

	if (isLoading) {
		return (
			<p className={styles["message"]} role="status">
				{AssetsPageMessage.TREE_LOADING}
			</p>
		);
	}

	if (isError) {
		return (
			<div className={styles["message"]}>
				<p className={styles["text"]}>{AssetsPageMessage.TREE_ERROR}</p>
				<Button
					label={AssetsPageMessage.TRY_AGAIN}
					onClick={handleRetryClick}
					size={ControlSize.SM}
					type="button"
					variant={ButtonVariant.SECONDARY}
				/>
			</div>
		);
	}

	return (
		<ul
			aria-label={AssetsPageMessage.TREE_LABEL}
			className={styles["tree"]}
			onFocus={handleFocus}
			onKeyDown={handleKeyDown}
			ref={treeReference}
			role="tree"
		>
			{data?.items.map((node) => (
				<TreeNode
					expandedAssetIds={expandedAssetIds}
					focusedAssetId={focusedAssetId}
					key={node.assetId}
					level={TREE_ROOT_LEVEL}
					node={node}
					onToggle={handleToggle}
					selectedAssetId={selectedAssetId}
				/>
			))}
		</ul>
	);
};

export { AssetTree };
