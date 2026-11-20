import { AssetsPageMessage } from "../../libs/enums/enums.js";
import { useAssetSearch } from "../../libs/hooks/use-asset-search/use-asset-search.hook.js";
import { AssetSearch } from "../asset-search/asset-search.js";
import { AssetTree } from "../asset-tree/asset-tree.js";
import styles from "./styles.module.css";

type Properties = {
	selectedAssetId: null | string;
};

/**
 * The sidebar's navigation: the search and, below it, the tree. While a
 * search is in the address its results take the tree's place (feature 003
 * FR-001).
 */
const AssetExplorer: React.FC<Properties> = ({
	selectedAssetId,
}: Properties) => {
	const { activeSearch, clearSearch, submitSearch } = useAssetSearch();

	return (
		<div className={styles["explorer"]}>
			<AssetSearch
				activeSearch={activeSearch}
				onClear={clearSearch}
				onSubmit={submitSearch}
			/>
			{!activeSearch && (
				<div className={styles["tree"]}>
					<p className={styles["label"]}>
						{AssetsPageMessage.SUBSTATIONS_LABEL}
					</p>
					<div className={styles["tree-scroll"]}>
						<AssetTree selectedAssetId={selectedAssetId} />
					</div>
				</div>
			)}
		</div>
	);
};

export { AssetExplorer };
