import {
	SkeletonShape,
	SkeletonWidth,
} from "~/libs/components/skeleton/libs/enums/enums.js";
import { Skeleton } from "~/libs/components/skeleton/skeleton.js";

import {
	DETAILS_SKELETON_ATTRIBUTES,
	DETAILS_SKELETON_CHILD_ROWS,
} from "../../../../libs/constants/constants.js";
import { AssetsPageMessage } from "../../../../libs/enums/enums.js";
import {
	createPositions,
	getSkeletonRowWidth,
} from "../../../../libs/helpers/helpers.js";
import styles from "./styles.module.css";

/**
 * Placeholders shaped like the details — the intro, the details card and the
 * children card — while an asset loads (spec 005 FR-019, FR-020, FR-022).
 */
const DetailsSkeleton: React.FC = () => (
	<div className={styles["skeleton"]} role="status">
		<span className="visually-hidden">{AssetsPageMessage.DETAILS_LOADING}</span>
		<div aria-hidden="true" className={styles["parts"]}>
			<div className={styles["intro"]} data-skeleton-part="intro">
				<Skeleton width={SkeletonWidth.QUARTER} />
				<Skeleton width={SkeletonWidth.QUARTER} />
				<Skeleton shape={SkeletonShape.TITLE} width={SkeletonWidth.HALF} />
				<Skeleton width={SkeletonWidth.QUARTER} />
			</div>
			<div className={styles["card"]}>
				<Skeleton width={SkeletonWidth.QUARTER} />
				<div className={styles["attributes"]}>
					{createPositions(DETAILS_SKELETON_ATTRIBUTES).map((position) => (
						<div
							className={styles["attribute"]}
							data-skeleton-part="attribute"
							key={position}
						>
							<Skeleton width={SkeletonWidth.HALF} />
							<Skeleton width={getSkeletonRowWidth(position)} />
						</div>
					))}
				</div>
			</div>
			<div className={styles["card"]}>
				<Skeleton width={SkeletonWidth.QUARTER} />
				{createPositions(DETAILS_SKELETON_CHILD_ROWS).map((position) => (
					<div
						className={styles["child-row"]}
						data-skeleton-part="child-row"
						key={position}
					>
						<Skeleton width={getSkeletonRowWidth(position)} />
					</div>
				))}
			</div>
		</div>
	</div>
);

export { DetailsSkeleton };
