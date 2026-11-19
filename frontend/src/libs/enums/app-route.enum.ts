const AppRoute = {
	ANY: "*",
	ASSETS: "/assets",
	ASSETS_$ASSET_ID: "/assets/:assetId",
	ROOT: "/",
} as const;

export { AppRoute };
