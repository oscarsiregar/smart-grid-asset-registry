import { generatePath } from "react-router-dom";

import { AppRoute } from "~/libs/enums/enums.js";

/** The asset's address, with the ID percent-encoded (research R2). */
const getAssetPath = (assetId: string): string =>
	generatePath(AppRoute.ASSETS_$ASSET_ID, { assetId });

export { getAssetPath };
