import { AssetDetails } from "../asset-details/asset-details.js";
import { AssetTree } from "../asset-tree/asset-tree.js";
import styles from "./styles.module.css";

type Properties = {
	selectedAssetId: null | string;
};

/** The tree beside the details area; stacked on narrow screens. */
const AssetExplorer: React.FC<Properties> = ({
	selectedAssetId,
}: Properties) => (
	<div className={styles["explorer"]}>
		<div className={styles["tree"]}>
			<AssetTree selectedAssetId={selectedAssetId} />
		</div>
		<div className={styles["details"]}>
			<AssetDetails selectedAssetId={selectedAssetId} />
		</div>
	</div>
);

export { AssetExplorer };
