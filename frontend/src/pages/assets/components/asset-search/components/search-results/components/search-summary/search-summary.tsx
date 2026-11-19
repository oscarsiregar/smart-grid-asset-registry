import { EMPTY_LENGTH } from "~/libs/constants/constants.js";
import { type AssetSearchRequestDto } from "~/modules/assets/assets.js";

import { AssetsPageMessage } from "../../../../../../libs/enums/enums.js";
import { getSearchSummaryText } from "../../../../../../libs/helpers/helpers.js";
import styles from "./styles.module.css";

type Properties = {
	receivedCount: number;
	search: AssetSearchRequestDto;
	totalItems: number;
};

/** How many assets match, in words; no match adds a hint (US3). */
const SearchSummary: React.FC<Properties> = ({
	receivedCount,
	search,
	totalItems,
}: Properties) => (
	<div className={styles["summary"]}>
		<p className={styles["text"]}>
			{getSearchSummaryText({ receivedCount, search, totalItems })}
		</p>
		{totalItems === EMPTY_LENGTH && (
			<p className={styles["hint"]}>{AssetsPageMessage.NO_MATCH_HINT}</p>
		)}
	</div>
);

export { SearchSummary };
