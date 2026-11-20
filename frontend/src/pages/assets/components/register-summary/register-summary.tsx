import {
	SkeletonShape,
	SkeletonWidth,
} from "~/libs/components/skeleton/libs/enums/enums.js";
import { Skeleton } from "~/libs/components/skeleton/skeleton.js";
import { EMPTY_LENGTH } from "~/libs/constants/constants.js";
import { useGetAssetNodesQuery } from "~/modules/assets/assets.js";

import { AssetsPageMessage } from "../../libs/enums/enums.js";
import { getRegisterUnitText } from "../../libs/helpers/helpers.js";
import styles from "./styles.module.css";

/**
 * The number of stored substations at the bottom of the sidebar (spec 005
 * FR-009), from the top-level list the tree reads — the same cached query, so
 * no extra request (research R4). An empty registry or a failed read shows
 * nothing; while it loads, the number is a placeholder.
 */
const RegisterSummary: React.FC = () => {
	const { data, isError } = useGetAssetNodesQuery(null);

	// The tree states a failure with its retry; no placeholder stays (spec 005
	// SC-004).
	if (isError || data?.totalItems === EMPTY_LENGTH) {
		return null;
	}

	return (
		<section className={styles["register"]}>
			<p className={styles["label"]}>{AssetsPageMessage.REGISTER_LABEL}</p>
			{data ? (
				<p className={styles["count"]}>
					<span className={styles["number"]}>{data.totalItems}</span>{" "}
					{getRegisterUnitText(data.totalItems)}
				</p>
			) : (
				<Skeleton shape={SkeletonShape.TITLE} width={SkeletonWidth.HALF} />
			)}
		</section>
	);
};

export { RegisterSummary };
