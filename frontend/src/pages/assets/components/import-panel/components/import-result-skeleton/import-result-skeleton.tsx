import {
	SkeletonShape,
	SkeletonWidth,
} from "~/libs/components/skeleton/libs/enums/enums.js";
import { Skeleton } from "~/libs/components/skeleton/skeleton.js";

import styles from "./styles.module.css";

// The result's parts: the five count cards and the first table rows.
const COUNT_CARDS = ["total", "imported", "unchanged", "rejected", "warnings"];
const TABLE_ROWS = ["first", "second", "third"];

/**
 * Placeholders for the banner, the counts and the first table rows while a
 * file is checked (spec 005 FR-019). Hidden from assistive technology: the
 * form already announces "Checking file…".
 */
const ImportResultSkeleton: React.FC = () => (
	<div aria-hidden="true" className={styles["skeleton"]} data-import-skeleton>
		<Skeleton shape={SkeletonShape.BLOCK} />
		<div className={styles["counts"]}>
			{COUNT_CARDS.map((card) => (
				<Skeleton key={card} shape={SkeletonShape.BLOCK} />
			))}
		</div>
		<div className={styles["rows"]}>
			{TABLE_ROWS.map((row) => (
				<Skeleton key={row} width={SkeletonWidth.THREE_QUARTERS} />
			))}
		</div>
	</div>
);

export { ImportResultSkeleton };
