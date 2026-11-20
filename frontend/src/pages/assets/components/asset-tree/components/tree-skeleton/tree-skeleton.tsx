import { SkeletonShape } from "~/libs/components/skeleton/libs/enums/enums.js";
import { Skeleton } from "~/libs/components/skeleton/skeleton.js";

import {
	createPositions,
	getSkeletonRowWidth,
} from "../../../../libs/helpers/helpers.js";
import styles from "./styles.module.css";

type Properties = {
	/** The loading text announced to assistive technology. */
	label: string;
	rows: number;
};

/**
 * Placeholder tree rows — a status shape and a name — while a list loads
 * (spec 005 FR-019, FR-022). Rows have a tree row's height, so nothing moves
 * when the nodes arrive (FR-020).
 */
const TreeSkeleton: React.FC<Properties> = ({ label, rows }: Properties) => (
	<div className={styles["skeleton"]} role="status">
		<span className="visually-hidden">{label}</span>
		{createPositions(rows).map((position) => (
			<div
				aria-hidden="true"
				className={styles["row"]}
				data-skeleton-row
				key={position}
			>
				<Skeleton shape={SkeletonShape.DOT} />
				<Skeleton width={getSkeletonRowWidth(position)} />
			</div>
		))}
	</div>
);

export { TreeSkeleton };
