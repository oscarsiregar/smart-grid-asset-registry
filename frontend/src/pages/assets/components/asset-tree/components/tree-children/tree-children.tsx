import { useCallback } from "react";

import { Button } from "~/libs/components/button/button.js";
import { ButtonVariant, ControlSize } from "~/libs/enums/enums.js";
import { useGetAssetNodesQuery } from "~/modules/assets/assets.js";

import { AssetsPageMessage } from "../../../../libs/enums/enums.js";
import { TreeNode } from "../tree-node/tree-node.js";
import styles from "./styles.module.css";

type Properties = {
	expandedAssetIds: ReadonlySet<string>;
	focusedAssetId: null | string;
	level: number;
	onToggle: (assetId: string) => void;
	parentAssetId: string;
	selectedAssetId: null | string;
};

/**
 * The children of an expanded node. Mounted only while the node is expanded,
 * so its query runs on the first expansion and is served from the cache
 * afterwards (research R7).
 */
const TreeChildren: React.FC<Properties> = ({
	expandedAssetIds,
	focusedAssetId,
	level,
	onToggle,
	parentAssetId,
	selectedAssetId,
}: Properties) => {
	const { data, isError, isLoading, refetch } =
		useGetAssetNodesQuery(parentAssetId);

	const handleRetryClick = useCallback((): void => {
		void refetch();
	}, [refetch]);

	return (
		<ul className={styles["group"]} role="group">
			{isLoading && (
				<li className={styles["message"]} role="none">
					<span role="status">{AssetsPageMessage.CHILDREN_LOADING}</span>
				</li>
			)}
			{isError && (
				<li className={styles["message"]} role="none">
					<span>{AssetsPageMessage.CHILDREN_ERROR}</span>
					<Button
						label={AssetsPageMessage.TRY_AGAIN}
						onClick={handleRetryClick}
						size={ControlSize.SM}
						type="button"
						variant={ButtonVariant.SECONDARY}
					/>
				</li>
			)}
			{data?.items.map((child) => (
				<TreeNode
					expandedAssetIds={expandedAssetIds}
					focusedAssetId={focusedAssetId}
					key={child.assetId}
					level={level}
					node={child}
					onToggle={onToggle}
					selectedAssetId={selectedAssetId}
				/>
			))}
		</ul>
	);
};

export { TreeChildren };
