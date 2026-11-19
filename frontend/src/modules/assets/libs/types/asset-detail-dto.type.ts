import { type AssetDto } from "./asset-dto.type.js";
import { type AssetNodeDto } from "./asset-node-dto.type.js";
import { type DescendantCountsDto } from "./descendant-counts-dto.type.js";

/**
 * `path` runs from the substation down to the asset itself;
 * `descendantCounts` is `null` unless the asset is a substation.
 */
type AssetDetailDto = {
	asset: AssetDto;
	children: AssetNodeDto[];
	descendantCounts: DescendantCountsDto | null;
	path: AssetNodeDto[];
};

export { type AssetDetailDto };
