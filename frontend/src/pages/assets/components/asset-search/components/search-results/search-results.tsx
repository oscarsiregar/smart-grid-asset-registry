import { configureString } from "~/libs/helpers/helpers.js";
import {
	type AssetSearchRequestDto,
	useSearchAssetsQuery,
} from "~/modules/assets/assets.js";

import { AssetsPageMessage } from "../../../../libs/enums/enums.js";
import { SearchResult } from "./components/search-result/search-result.js";
import styles from "./styles.module.css";

type Properties = {
	search: AssetSearchRequestDto;
};

// One match reads "1 asset matches", every other count "N assets match".
const SINGLE_MATCH = 1;

const getSummaryText = (totalItems: number, text: string): string =>
	configureString(
		totalItems === SINGLE_MATCH
			? AssetsPageMessage.RESULTS_SUMMARY_ONE
			: AssetsPageMessage.RESULTS_SUMMARY,
		{ count: String(totalItems), text },
	);

/**
 * The results of the search in the address, in place of the tree (FR-001).
 * Matching, order, the page and the total are the backend's; the list is
 * shown as received (principle II).
 */
const SearchResults: React.FC<Properties> = ({ search }: Properties) => {
	const { currentData } = useSearchAssetsQuery(search);

	const summary =
		currentData === undefined
			? null
			: getSummaryText(currentData.totalItems, search.q);

	return (
		<section
			aria-label={AssetsPageMessage.RESULTS_LABEL}
			className={styles["results"]}
		>
			{summary && <p className={styles["summary"]}>{summary}</p>}
			{currentData && (
				<ul className={styles["list"]}>
					{currentData.items.map((result) => (
						<SearchResult key={result.assetId} result={result} />
					))}
				</ul>
			)}
		</section>
	);
};

export { SearchResults };
