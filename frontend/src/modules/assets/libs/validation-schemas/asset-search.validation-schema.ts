import { z } from "zod";

import { SEARCH_TEXT_MIN_LENGTH } from "../constants/constants.js";
import { AssetSearchValidationMessage, AssetType } from "../enums/enums.js";

/**
 * A search: the text (trimmed) and an asset type, or `null` for all types.
 * The form only refuses a blank text (spec FR-004); the length limit and
 * every other rule are the backend's (principle II), and its message is shown
 * when it refuses a search.
 */
const assetSearch = z.object({
	assetType: z.enum(AssetType).nullable(),
	q: z.string().trim().min(SEARCH_TEXT_MIN_LENGTH, {
		error: AssetSearchValidationMessage.TEXT_REQUIRED,
	}),
});

export { assetSearch };
