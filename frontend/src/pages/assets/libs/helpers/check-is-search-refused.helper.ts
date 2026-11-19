import { checkIsServerError } from "~/libs/modules/api/libs/helpers/check-is-server-error.helper.js";
import { AssetsErrorCode } from "~/modules/assets/assets.js";

/**
 * The backend refused the search (400 INVALID_REQUEST, for example a text
 * that is too long): its message is shown, and retrying would be refused again.
 */
const checkIsSearchRefused = (error: unknown): boolean =>
	checkIsServerError(error) && error.code === AssetsErrorCode.INVALID_REQUEST;

export { checkIsSearchRefused };
