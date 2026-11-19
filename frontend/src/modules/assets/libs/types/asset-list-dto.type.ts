import { type AssetNodeDto } from "./asset-node-dto.type.js";

/** A complete, ordered list; never paged (Template Exception 13). */
type AssetListDto = {
	items: AssetNodeDto[];
	totalItems: number;
};

export { type AssetListDto };
