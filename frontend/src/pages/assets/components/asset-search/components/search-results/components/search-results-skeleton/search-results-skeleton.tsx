import { SkeletonWidth } from "~/libs/components/skeleton/libs/enums/enums.js";
import { Skeleton } from "~/libs/components/skeleton/skeleton.js";

import { SEARCH_SKELETON_ROWS } from "../../../../../../libs/constants/constants.js";
import { AssetsPageMessage } from "../../../../../../libs/enums/enums.js";
import {
	createPositions,
	getSkeletonRowWidth,
} from "../../../../../../libs/helpers/helpers.js";
import styles from "./styles.module.css";

/**
 * Placeholder result rows — the ID and name line, then the type, status and
 * location line — while a search runs (spec 005 FR-019, FR-022).
 */
const SearchResultsSkeleton: React.FC = () => (
	<div className={styles["skeleton"]} role="status">
		<span className="visually-hidden">{AssetsPageMessage.SEARCHING}</span>
		{createPositions(SEARCH_SKELETON_ROWS).map((position) => (
			<div
				aria-hidden="true"
				className={styles["row"]}
				data-skeleton-row
				key={position}
			>
				<Skeleton width={getSkeletonRowWidth(position)} />
				<Skeleton width={SkeletonWidth.HALF} />
			</div>
		))}
	</div>
);

export { SearchResultsSkeleton };
