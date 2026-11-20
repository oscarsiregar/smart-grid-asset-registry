import { useCallback } from "react";

import { Button } from "~/libs/components/button/button.js";
import { EMPTY_LENGTH } from "~/libs/constants/constants.js";
import { ButtonVariant, ControlSize } from "~/libs/enums/enums.js";
import { getErrorMessage } from "~/libs/modules/api/libs/helpers/get-error-message.helper.js";
import {
	type AssetSearchRequestDto,
	useSearchAssetsQuery,
} from "~/modules/assets/assets.js";

import { AssetsPageMessage } from "../../../../libs/enums/enums.js";
import { checkIsSearchRefused } from "../../../../libs/helpers/helpers.js";
import { SearchResult } from "./components/search-result/search-result.js";
import { SearchResultsSkeleton } from "./components/search-results-skeleton/search-results-skeleton.js";
import { SearchSummary } from "./components/search-summary/search-summary.js";
import styles from "./styles.module.css";

type Properties = {
	search: AssetSearchRequestDto;
};

/**
 * The results of the search in the address, in place of the tree (FR-001),
 * and every state of it in words (US3). Matching, order, the page and the
 * total are the backend's; the list is shown as received (principle II).
 */
const SearchResults: React.FC<Properties> = ({ search }: Properties) => {
	const { currentData, error, isError, refetch } = useSearchAssetsQuery(search);

	const handleRetryClick = useCallback((): void => {
		void refetch();
	}, [refetch]);

	const renderContent = (): React.ReactNode => {
		// A refused search would be refused again: its message, no retry.
		if (checkIsSearchRefused(error)) {
			return (
				<p className={styles["message"]}>
					{getErrorMessage(error, AssetsPageMessage.SEARCH_ERROR)}
				</p>
			);
		}

		if (isError) {
			return (
				<div className={styles["failure"]}>
					<p className={styles["message"]}>{AssetsPageMessage.SEARCH_ERROR}</p>
					<Button
						label={AssetsPageMessage.TRY_AGAIN}
						onClick={handleRetryClick}
						size={ControlSize.SM}
						type="button"
						variant={ButtonVariant.SECONDARY}
					/>
				</div>
			);
		}

		if (!currentData) {
			return <SearchResultsSkeleton />;
		}

		return (
			<>
				<SearchSummary
					receivedCount={currentData.items.length}
					search={search}
					totalItems={currentData.totalItems}
				/>
				{currentData.items.length > EMPTY_LENGTH && (
					<ul className={styles["list"]}>
						{currentData.items.map((result) => (
							<SearchResult key={result.assetId} result={result} />
						))}
					</ul>
				)}
			</>
		);
	};

	return (
		<section
			aria-label={AssetsPageMessage.RESULTS_LABEL}
			className={styles["results"]}
		>
			{renderContent()}
		</section>
	);
};

export { SearchResults };
