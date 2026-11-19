import { type AssetSearchItemDto } from "./asset-search-item-dto.type.js";

/** One page of matches (the backend's first 50); `totalItems` counts every match. */
type AssetSearchResultDto = {
	items: AssetSearchItemDto[];
	totalItems: number;
};

export { type AssetSearchResultDto };
