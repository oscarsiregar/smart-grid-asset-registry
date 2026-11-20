import { skipToken } from "@reduxjs/toolkit/query/react";

import { useGetAssetDetailQuery } from "~/modules/assets/assets.js";

import { AssetsPageMessage } from "../../libs/enums/enums.js";
import styles from "./styles.module.css";

type Properties = {
	selectedAssetId: null | string;
};

/**
 * The page title and, under it, the selected asset's path names as one line
 * of text (spec 005 FR-010, research R6). It reads the details query the
 * details area uses, so it makes no extra request; without loaded details it
 * shows the page description.
 */
const AssetHeader: React.FC<Properties> = ({ selectedAssetId }: Properties) => {
	const { currentData } = useGetAssetDetailQuery(selectedAssetId ?? skipToken);
	const line = currentData
		? currentData.path
				.map((step) => step.assetName)
				.join(AssetsPageMessage.PATH_SEPARATOR)
		: AssetsPageMessage.DESCRIPTION;

	return (
		<div className={styles["header"]}>
			<h1 className={styles["title"]}>{AssetsPageMessage.TITLE}</h1>
			<p className={styles["line"]} title={line}>
				{line}
			</p>
		</div>
	);
};

export { AssetHeader };
