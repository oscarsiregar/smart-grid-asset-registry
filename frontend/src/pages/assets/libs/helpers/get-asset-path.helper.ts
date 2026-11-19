import { AppRoute } from "~/libs/enums/enums.js";
import { generateRoutePath } from "~/libs/helpers/helpers.js";
import { type GeneratedRoutePath } from "~/libs/types/types.js";

/** The asset's address, with the ID percent-encoded (research R2). */
const getAssetPath = (assetId: string): GeneratedRoutePath =>
	generateRoutePath(AppRoute.ASSETS_$ASSET_ID, { assetId });

export { getAssetPath };
