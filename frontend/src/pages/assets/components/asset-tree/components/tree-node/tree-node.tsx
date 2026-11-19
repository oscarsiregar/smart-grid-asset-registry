import { useCallback } from "react";

import { Icon } from "~/libs/components/icon/icon.js";
import { Link } from "~/libs/components/link/link.js";
import { EMPTY_LENGTH } from "~/libs/constants/constants.js";
import { IconName } from "~/libs/enums/enums.js";
import { configureString, getValidClasses } from "~/libs/helpers/helpers.js";
import { type AssetNodeDto } from "~/modules/assets/assets.js";

import { TREE_STEP } from "../../../../libs/constants/constants.js";
import { AssetsPageMessage, TabIndex } from "../../../../libs/enums/enums.js";
import { getAssetPath } from "../../../../libs/helpers/helpers.js";
import { assetTypeToLabel } from "../../../../libs/maps.js";
import { StatusLabel } from "../status-label/status-label.js";
import { TreeChildren } from "../tree-children/tree-children.js";
import styles from "./styles.module.css";

type Properties = {
	expandedAssetIds: ReadonlySet<string>;
	focusedAssetId: null | string;
	level: number;
	node: AssetNodeDto;
	onToggle: (assetId: string) => void;
	selectedAssetId: null | string;
};

/**
 * One node of the navigation tree (research R8): the label is a link, so
 * selecting is navigating; a separate toggle expands it. Only nodes with
 * children are expandable (spec FR-003).
 */
const TreeNode: React.FC<Properties> = ({
	expandedAssetIds,
	focusedAssetId,
	level,
	node,
	onToggle,
	selectedAssetId,
}: Properties) => {
	const isExpandable = node.childCount > EMPTY_LENGTH;
	const isExpanded = isExpandable && expandedAssetIds.has(node.assetId);

	const handleToggleClick = useCallback((): void => {
		onToggle(node.assetId);
	}, [node.assetId, onToggle]);

	const toggleLabel = configureString(
		isExpanded
			? AssetsPageMessage.COLLAPSE_LABEL
			: AssetsPageMessage.EXPAND_LABEL,
		{ assetId: node.assetId },
	);

	return (
		<li
			aria-expanded={isExpandable ? isExpanded : undefined}
			aria-level={level}
			aria-selected={node.assetId === selectedAssetId}
			className={styles["node"]}
			data-asset-id={node.assetId}
			role="treeitem"
			tabIndex={
				node.assetId === focusedAssetId ? TabIndex.FOCUSABLE : TabIndex.SKIPPED
			}
		>
			<div className={styles["row"]}>
				{isExpandable ? (
					<button
						aria-label={toggleLabel}
						className={styles["toggle"]}
						onClick={handleToggleClick}
						tabIndex={TabIndex.SKIPPED}
						type="button"
					>
						<Icon
							className={getValidClasses(
								styles["chevron"],
								isExpanded && styles["expanded"],
							)}
							iconName={IconName.CHEVRON}
						/>
					</button>
				) : (
					<span className={styles["spacer"]} />
				)}
				<Link
					className={styles["label"]}
					hasDefaultStyles={false}
					tabIndex={TabIndex.SKIPPED}
					to={getAssetPath(node.assetId)}
				>
					<span className={styles["id"]}>{node.assetId}</span>
					<span className={styles["name"]}>{node.assetName}</span>
					<span className={styles["type"]}>
						{assetTypeToLabel[node.assetType]}
					</span>
				</Link>
				<StatusLabel status={node.operationalStatus} />
			</div>
			{isExpanded && (
				<TreeChildren
					expandedAssetIds={expandedAssetIds}
					focusedAssetId={focusedAssetId}
					level={level + TREE_STEP}
					onToggle={onToggle}
					parentAssetId={node.assetId}
					selectedAssetId={selectedAssetId}
				/>
			)}
		</li>
	);
};

export { TreeNode };
