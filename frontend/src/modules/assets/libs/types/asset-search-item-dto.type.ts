import { type AssetNodeDto } from "./asset-node-dto.type.js";

/** A search match: the asset node plus its ancestors, substation first ([] for a substation). */
type AssetSearchItemDto = AssetNodeDto & {
	ancestorAssetIds: string[];
};

export { type AssetSearchItemDto };
