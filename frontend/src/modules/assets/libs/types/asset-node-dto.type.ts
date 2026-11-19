import { type ValueOf } from "~/libs/types/types.js";

import { type AssetType, type OperationalStatus } from "../enums/enums.js";

/** A list item or path step; `childCount > 0` makes it expandable. */
type AssetNodeDto = {
	assetId: string;
	assetName: string;
	assetType: ValueOf<typeof AssetType>;
	childCount: number;
	operationalStatus: null | ValueOf<typeof OperationalStatus>;
};

export { type AssetNodeDto };
