import { type ValueOf } from "~/libs/types/types.js";
import { type AssetType } from "~/modules/assets/assets.js";

/** One group of a substation's children in the details area (spec FR-012). */
type ChildGroup = {
	assetType: ValueOf<typeof AssetType>;
	label: string;
};

export { type ChildGroup };
