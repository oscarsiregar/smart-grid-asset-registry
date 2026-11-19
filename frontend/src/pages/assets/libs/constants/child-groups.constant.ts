import { AssetType } from "~/modules/assets/assets.js";

import { AssetsPageMessage } from "../enums/enums.js";
import { type ChildGroup } from "../types/types.js";

/** A substation's child groups, in the order of spec FR-012. */
const CHILD_GROUPS: readonly ChildGroup[] = [
	{
		assetType: AssetType.TRANSFORMER,
		label: AssetsPageMessage.COUNT_TRANSFORMERS,
	},
	{ assetType: AssetType.LV_BOARD, label: AssetsPageMessage.COUNT_LV_BOARDS },
	{
		assetType: AssetType.SWITCHBOARD,
		label: AssetsPageMessage.COUNT_SWITCHBOARDS,
	},
];

export { CHILD_GROUPS };
