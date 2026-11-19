import { skipToken } from "@reduxjs/toolkit/query/react";
import { useCallback } from "react";

import { useGetAssetDetailQuery } from "~/modules/assets/assets.js";

import { AssetsPageMessage } from "../../libs/enums/enums.js";
import { assetTypeToLabel } from "../../libs/maps.js";
import { AssetAttributes } from "./components/asset-attributes/asset-attributes.js";
import { AssetPath } from "./components/asset-path/asset-path.js";
import { DetailsMessage } from "./components/details-message/details-message.js";
import styles from "./styles.module.css";

type Properties = {
	selectedAssetId: null | string;
};

/**
 * The details area: a prompt without a selection, otherwise the selected
 * asset's path and attributes (spec FR-008 – FR-011, FR-016). `currentData`
 * is the data of the current selection only, so switching assets shows the
 * loading state instead of the previous asset.
 */
const AssetDetails: React.FC<Properties> = ({
	selectedAssetId,
}: Properties) => {
	const { currentData, isError, refetch } = useGetAssetDetailQuery(
		selectedAssetId ?? skipToken,
	);

	const handleRetryClick = useCallback((): void => {
		void refetch();
	}, [refetch]);

	if (selectedAssetId === null) {
		return <DetailsMessage message={AssetsPageMessage.SELECT_PROMPT} />;
	}

	if (isError) {
		return (
			<DetailsMessage
				message={AssetsPageMessage.DETAILS_ERROR}
				onRetry={handleRetryClick}
			/>
		);
	}

	if (!currentData) {
		return (
			<DetailsMessage isLoading message={AssetsPageMessage.DETAILS_LOADING} />
		);
	}

	const { asset, path } = currentData;
	const parent =
		path.find((step) => step.assetId === asset.parentAssetId) ?? null;

	return (
		<article className={styles["details"]}>
			<AssetPath currentAssetId={asset.assetId} path={path} />
			<header className={styles["header"]}>
				<h3 className={styles["title"]}>{asset.assetName}</h3>
				<p className={styles["type"]}>{assetTypeToLabel[asset.assetType]}</p>
			</header>
			<AssetAttributes asset={asset} parent={parent} />
		</article>
	);
};

export { AssetDetails };
