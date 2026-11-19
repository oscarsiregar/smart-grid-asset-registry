import { UNADDRESSABLE_ASSET_IDS } from "../constants/constants.js";

/** Whether the asset can have an address, and so be selected and opened. */
const checkIsAddressableAssetId = (assetId: string): boolean =>
	!UNADDRESSABLE_ASSET_IDS.includes(assetId);

export { checkIsAddressableAssetId };
