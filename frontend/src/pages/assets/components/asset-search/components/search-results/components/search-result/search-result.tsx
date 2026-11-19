import { Link } from "~/libs/components/link/link.js";
import { type AssetSearchItemDto } from "~/modules/assets/assets.js";

import { AssetsPageMessage } from "../../../../../../libs/enums/enums.js";
import {
	checkIsAddressableAssetId,
	getAssetLocationText,
	getAssetPath,
} from "../../../../../../libs/helpers/helpers.js";
import { assetTypeToLabel } from "../../../../../../libs/maps.js";
import { StatusBadge } from "../../../../../status-badge/status-badge.js";
import styles from "./styles.module.css";

type Properties = {
	result: AssetSearchItemDto;
};

/**
 * One search result. Its link is the asset's address, so opening it is
 * feature 002's deep link (FR-011). IDs `.` and `..` cannot have an address
 * and are plain text (FR-012).
 */
const SearchResult: React.FC<Properties> = ({ result }: Properties) => {
	const label = (
		<>
			<span className={styles["id"]}>{result.assetId}</span>
			<span className={styles["name"]}>{result.assetName}</span>
		</>
	);

	return (
		<li className={styles["result"]} data-asset-id={result.assetId}>
			{checkIsAddressableAssetId(result.assetId) ? (
				<Link className={styles["label"]} to={getAssetPath(result.assetId)}>
					{label}
				</Link>
			) : (
				<span className={styles["label"]}>{label}</span>
			)}
			<span className={styles["facts"]}>
				<span>{assetTypeToLabel[result.assetType]}</span>
				<StatusBadge isCompact status={result.operationalStatus} />
				<span className={styles["location"]}>
					{getAssetLocationText(result.ancestorAssetIds)}
				</span>
			</span>
			{!checkIsAddressableAssetId(result.assetId) && (
				<span className={styles["note"]}>{AssetsPageMessage.CANNOT_OPEN}</span>
			)}
		</li>
	);
};

export { SearchResult };
