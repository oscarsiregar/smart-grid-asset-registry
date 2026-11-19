import { useAssetSearch } from "../../libs/hooks/use-asset-search/use-asset-search.hook.js";
import { AssetDetails } from "../asset-details/asset-details.js";
import { AssetSearch } from "../asset-search/asset-search.js";
import { AssetTree } from "../asset-tree/asset-tree.js";
import styles from "./styles.module.css";

type Properties = {
	selectedAssetId: null | string;
};

/**
 * The search and the tree beside the details area; stacked on narrow
 * screens. While a search is in the address its results take the tree's
 * place (feature 003 FR-001); the details area is unchanged.
 */
const AssetExplorer: React.FC<Properties> = ({
	selectedAssetId,
}: Properties) => {
	const { activeSearch, clearSearch, submitSearch } = useAssetSearch();

	return (
		<div className={styles["explorer"]}>
			<div className={styles["tree"]}>
				<AssetSearch
					activeSearch={activeSearch}
					onClear={clearSearch}
					onSubmit={submitSearch}
				/>
				{!activeSearch && <AssetTree selectedAssetId={selectedAssetId} />}
			</div>
			<div className={styles["details"]}>
				<AssetDetails selectedAssetId={selectedAssetId} />
			</div>
		</div>
	);
};

export { AssetExplorer };
