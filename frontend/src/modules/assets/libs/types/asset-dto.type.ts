import { type ValueOf } from "~/libs/types/types.js";

import { type AssetType, type OperationalStatus } from "../enums/enums.js";

/** Every stored attribute; blank CSV values are `null`; the date is `YYYY-MM-DD`. */
type AssetDto = {
	assetId: string;
	assetName: string;
	assetType: ValueOf<typeof AssetType>;
	commissionedDate: null | string;
	manufacturer: null | string;
	model: null | string;
	operationalStatus: null | ValueOf<typeof OperationalStatus>;
	parentAssetId: null | string;
	ratingKva: null | number;
	serialNumber: null | string;
	voltageKv: null | number;
};

export { type AssetDto };
