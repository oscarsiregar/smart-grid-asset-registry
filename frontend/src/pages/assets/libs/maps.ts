import { type ValueOf } from "~/libs/types/types.js";
import { AssetType, OperationalStatus } from "~/modules/assets/assets.js";

import { AssetsPageMessage } from "./enums/enums.js";

const assetTypeToLabel = {
	[AssetType.LV_BOARD]: AssetsPageMessage.TYPE_LV_BOARD,
	[AssetType.SUBSTATION]: AssetsPageMessage.TYPE_SUBSTATION,
	[AssetType.SWITCHBOARD]: AssetsPageMessage.TYPE_SWITCHBOARD,
	[AssetType.SWITCHBOARD_PANEL]: AssetsPageMessage.TYPE_SWITCHBOARD_PANEL,
	[AssetType.TRANSFORMER]: AssetsPageMessage.TYPE_TRANSFORMER,
} satisfies Record<ValueOf<typeof AssetType>, string>;

const operationalStatusToLabel = {
	[OperationalStatus.IN_SERVICE]: AssetsPageMessage.STATUS_IN_SERVICE,
	[OperationalStatus.MAINTENANCE]: AssetsPageMessage.STATUS_MAINTENANCE,
	[OperationalStatus.OUT_OF_SERVICE]: AssetsPageMessage.STATUS_OUT_OF_SERVICE,
} satisfies Record<ValueOf<typeof OperationalStatus>, string>;

export { assetTypeToLabel, operationalStatusToLabel };
