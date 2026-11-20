import { useCallback, useEffect, useRef } from "react";

import { Icon } from "~/libs/components/icon/icon.js";
import { Link } from "~/libs/components/link/link.js";
import { EMPTY_LENGTH } from "~/libs/constants/constants.js";
import { IconName } from "~/libs/enums/enums.js";
import { configureString, getValidClasses } from "~/libs/helpers/helpers.js";
import { type AssetNodeDto } from "~/modules/assets/assets.js";

import {
	NODE_DETAIL_SEPARATOR,
	TREE_STEP,
} from "../../../../libs/constants/constants.js";
import { AssetsPageMessage, TabIndex } from "../../../../libs/enums/enums.js";
import {
	checkIsAddressableAssetId,
	getAssetPath,
	getStatusText,
} from "../../../../libs/helpers/helpers.js";
import { assetTypeToLabel } from "../../../../libs/maps.js";
import { StatusBadge } from "../../../status-badge/status-badge.js";
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
 * children are expandable (spec FR-003). The label shows the status shape and
 * the name; the ID, type and status in words are hidden text and the hover
 * title (spec 005 FR-007). The selected node is marked by an accent bar and
 * bold text, "Selected" for screen readers, and scrolled into view (spec
 * FR-006, spec 005 FR-008).
 */
const TreeNode: React.FC<Properties> = ({
	expandedAssetIds,
	focusedAssetId,
	level,
	node,
	onToggle,
	selectedAssetId,
}: Properties) => {
	const nodeReference = useRef<HTMLLIElement>(null);
	const isExpandable = node.childCount > EMPTY_LENGTH;
	const isExpanded = isExpandable && expandedAssetIds.has(node.assetId);
	const isSelected = node.assetId === selectedAssetId;

	// Only when the node becomes selected (a click, a link, a deep link once
	// its ancestors are expanded), not on every render.
	useEffect(() => {
		if (isSelected) {
			nodeReference.current?.scrollIntoView({ block: "nearest" });
		}
	}, [isSelected]);

	const handleToggleClick = useCallback((): void => {
		onToggle(node.assetId);
	}, [node.assetId, onToggle]);

	const toggleLabel = configureString(
		isExpanded
			? AssetsPageMessage.COLLAPSE_LABEL
			: AssetsPageMessage.EXPAND_LABEL,
		{ assetId: node.assetId },
	);

	const typeText = assetTypeToLabel[node.assetType];
	const statusText = getStatusText(node.operationalStatus);
	const hoverText = [node.assetName, node.assetId, typeText, statusText].join(
		NODE_DETAIL_SEPARATOR,
	);

	const labelContent = (
		<>
			<StatusBadge isShapeOnly status={node.operationalStatus} />
			<span className={styles["name"]}>{node.assetName}</span>
			<span className="visually-hidden">{node.assetId}</span>
			<span className="visually-hidden">{typeText}</span>
			<span className="visually-hidden">{statusText}</span>
			{isSelected && (
				<span className="visually-hidden">
					{AssetsPageMessage.SELECTED_MARKER}
				</span>
			)}
		</>
	);

	return (
		<li
			aria-expanded={isExpandable ? isExpanded : undefined}
			aria-level={level}
			aria-selected={isSelected}
			className={styles["node"]}
			data-asset-id={node.assetId}
			ref={nodeReference}
			role="treeitem"
			tabIndex={
				node.assetId === focusedAssetId ? TabIndex.FOCUSABLE : TabIndex.SKIPPED
			}
		>
			<div
				className={getValidClasses(
					styles["row"],
					isSelected && styles["selected"],
				)}
			>
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
				{checkIsAddressableAssetId(node.assetId) ? (
					<Link
						className={styles["label"]}
						hasDefaultStyles={false}
						tabIndex={TabIndex.SKIPPED}
						title={hoverText}
						to={getAssetPath(node.assetId)}
					>
						{labelContent}
					</Link>
				) : (
					<span className={styles["label"]} title={hoverText}>
						{labelContent}
					</span>
				)}
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
