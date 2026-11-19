import { EMPTY_LENGTH } from "~/libs/constants/constants.js";

import { LOCATION_SEPARATOR } from "../constants/constants.js";
import { AssetsPageMessage } from "../enums/enums.js";

/** Where a search result sits: its ancestors, substation first, or the top level (spec FR-007). */
const getAssetLocationText = (ancestorAssetIds: readonly string[]): string =>
	ancestorAssetIds.length === EMPTY_LENGTH
		? AssetsPageMessage.TOP_LEVEL
		: ancestorAssetIds.join(LOCATION_SEPARATOR);

export { getAssetLocationText };
