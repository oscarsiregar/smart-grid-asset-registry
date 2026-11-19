import { checkIsServerError } from "~/libs/modules/api/libs/helpers/check-is-server-error.helper.js";
import { AssetsErrorCode } from "~/modules/assets/assets.js";

/** The server has no asset with the requested ID: a state, not an error (spec FR-020). */
const checkIsAssetNotFound = (error: unknown): boolean =>
	checkIsServerError(error) && error.code === AssetsErrorCode.ASSET_NOT_FOUND;

export { checkIsAssetNotFound };
