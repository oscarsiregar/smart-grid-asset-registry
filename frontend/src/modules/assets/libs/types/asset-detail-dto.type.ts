import { type AssetDto } from "./asset-dto.type.js";
import { type AssetNodeDto } from "./asset-node-dto.type.js";

/** `path` runs from the substation down to the asset itself. */
type AssetDetailDto = {
	asset: AssetDto;
	children: AssetNodeDto[];
	path: AssetNodeDto[];
};

export { type AssetDetailDto };
