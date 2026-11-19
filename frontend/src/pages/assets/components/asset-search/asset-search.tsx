import { type AssetSearchRequestDto } from "~/modules/assets/assets.js";

import { SearchForm } from "./components/search-form/search-form.js";
import { SearchResults } from "./components/search-results/search-results.js";
import styles from "./styles.module.css";

type Properties = {
	activeSearch: AssetSearchRequestDto | null;
	onClear: () => void;
	onSubmit: (search: AssetSearchRequestDto) => void;
};

/** The search form and, while a search is in the address, its results. */
const AssetSearch: React.FC<Properties> = ({
	activeSearch,
	onClear,
	onSubmit,
}: Properties) => (
	<div className={styles["search"]}>
		<SearchForm
			activeSearch={activeSearch}
			// A new search in the address (Back, Forward, a link) refills the form.
			key={`${activeSearch?.q ?? ""}|${activeSearch?.assetType ?? ""}`}
			onClear={onClear}
			onSubmit={onSubmit}
		/>
		{activeSearch && <SearchResults search={activeSearch} />}
	</div>
);

export { AssetSearch };
