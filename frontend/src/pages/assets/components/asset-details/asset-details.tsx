import { skipToken } from "@reduxjs/toolkit/query/react";
import { useCallback } from "react";

import { configureString } from "~/libs/helpers/helpers.js";
import { AssetType, useGetAssetDetailQuery } from "~/modules/assets/assets.js";

import { AssetsPageMessage } from "../../libs/enums/enums.js";
import { checkIsAssetNotFound } from "../../libs/helpers/helpers.js";
import { assetTypeToLabel } from "../../libs/maps.js";
import { AssetAttributes } from "./components/asset-attributes/asset-attributes.js";
import { AssetPath } from "./components/asset-path/asset-path.js";
import { ChildGroups } from "./components/child-groups/child-groups.js";
import { ChildList } from "./components/child-list/child-list.js";
import { DescendantCounts } from "./components/descendant-counts/descendant-counts.js";
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
	const { currentData, error, isError, refetch } = useGetAssetDetailQuery(
		selectedAssetId ?? skipToken,
	);

	const handleRetryClick = useCallback((): void => {
		void refetch();
	}, [refetch]);

	if (selectedAssetId === null) {
		return <DetailsMessage message={AssetsPageMessage.SELECT_PROMPT} />;
	}

	// An unknown ID is a state of its own, not an error (spec FR-020).
	if (checkIsAssetNotFound(error)) {
		return (
			<DetailsMessage
				hasBackLink
				message={configureString(AssetsPageMessage.NOT_FOUND, {
					assetId: selectedAssetId,
				})}
			/>
		);
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

	const { asset, children, descendantCounts, path } = currentData;
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
			{asset.assetType === AssetType.SUBSTATION && (
				<ChildGroups childAssets={children} />
			)}
			{descendantCounts && <DescendantCounts counts={descendantCounts} />}
			{asset.assetType === AssetType.SWITCHBOARD && (
				<ChildList
					childAssets={children}
					label={AssetsPageMessage.COUNT_SWITCHBOARD_PANELS}
				/>
			)}
		</article>
	);
};

export { AssetDetails };
