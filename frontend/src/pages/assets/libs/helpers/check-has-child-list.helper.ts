import { type ValueOf } from "~/libs/types/types.js";
import { AssetType } from "~/modules/assets/assets.js";

/**
 * Whether the details list an asset's children: a substation's groups and a
 * switchboard's panels (feature 002 FR-012, FR-015). Other types have none,
 * so their children card says nothing is connected below (spec 005 FR-015).
 */
const checkHasChildList = (assetType: ValueOf<typeof AssetType>): boolean =>
	assetType === AssetType.SUBSTATION || assetType === AssetType.SWITCHBOARD;

export { checkHasChildList };
