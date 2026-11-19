import { type AssetSearchRequestDto } from "~/modules/assets/assets.js";

/** The form's values when the address holds no search. */
const EMPTY_SEARCH: AssetSearchRequestDto = { assetType: null, q: "" };

export { EMPTY_SEARCH };
