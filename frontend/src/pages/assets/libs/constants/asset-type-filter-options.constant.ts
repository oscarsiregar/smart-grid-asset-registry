import { type ValueOf } from "~/libs/types/types.js";
import { AssetType } from "~/modules/assets/assets.js";

/** The type field's options, in the hierarchy's order (spec FR-005), not the enum's key order. */
const ASSET_TYPE_FILTER_OPTIONS: readonly ValueOf<typeof AssetType>[] = [
	AssetType.SUBSTATION,
	AssetType.TRANSFORMER,
	AssetType.LV_BOARD,
	AssetType.SWITCHBOARD,
	AssetType.SWITCHBOARD_PANEL,
];

export { ASSET_TYPE_FILTER_OPTIONS };
