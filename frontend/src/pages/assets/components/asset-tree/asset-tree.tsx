import { skipToken } from "@reduxjs/toolkit/query/react";
import { useCallback, useMemo, useRef } from "react";

import { Button } from "~/libs/components/button/button.js";
import { ButtonVariant, ControlSize } from "~/libs/enums/enums.js";
import {
	useGetAssetDetailQuery,
	useGetAssetNodesQuery,
} from "~/modules/assets/assets.js";

import { TREE_ROOT_LEVEL } from "../../libs/constants/constants.js";
import { AssetsPageMessage } from "../../libs/enums/enums.js";
import { useExpandedNodes } from "../../libs/hooks/use-expanded-nodes/use-expanded-nodes.hook.js";
import { useTreeKeyboard } from "../../libs/hooks/use-tree-keyboard/use-tree-keyboard.hook.js";
import { TreeNode } from "./components/tree-node/tree-node.js";
import styles from "./styles.module.css";

type Properties = {
	selectedAssetId: null | string;
};

/**
 * The asset hierarchy as a WAI-ARIA navigation tree (research R8): every
 * substation at the top level, children loaded when a node is expanded.
 * Which nodes are expanded is local state, not kept across refresh; the
 * selected asset's ancestors are expanded from its path (spec FR-005,
 * FR-019), read from the same cached detail query the details area uses.
 */
const AssetTree: React.FC<Properties> = ({ selectedAssetId }: Properties) => {
	const { data, isError, isLoading, refetch } = useGetAssetNodesQuery(null);
	const treeReference = useRef<HTMLUListElement>(null);
	const { currentData: selectedDetail } = useGetAssetDetailQuery(
		selectedAssetId ?? skipToken,
	);
	const selectedPath = useMemo(
		() => selectedDetail?.path.map((step) => step.assetId) ?? null,
		[selectedDetail],
	);
	const { collapse, expand, expandedAssetIds, toggle } = useExpandedNodes({
		selectedAssetId,
		selectedPath,
	});

	const { focusedAssetId, handleFocus, handleKeyDown } = useTreeKeyboard({
		onCollapse: collapse,
		onExpand: expand,
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
					onToggle={toggle}
					selectedAssetId={selectedAssetId}
				/>
			))}
		</ul>
	);
};

export { AssetTree };
