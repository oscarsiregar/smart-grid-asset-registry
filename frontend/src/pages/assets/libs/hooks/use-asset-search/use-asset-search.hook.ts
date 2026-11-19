import { useCallback, useMemo } from "react";
import { useSearchParams } from "react-router-dom";

import { SearchParameter } from "~/libs/enums/enums.js";
import { type ValueOf } from "~/libs/types/types.js";
import {
	type AssetSearchRequestDto,
	AssetType,
} from "~/modules/assets/assets.js";

type AssetSearch = {
	/** The search in the address, or `null` when there is none (blank text). */
	activeSearch: AssetSearchRequestDto | null;
	clearSearch: () => void;
	submitSearch: (search: AssetSearchRequestDto) => void;
};

const ASSET_TYPES: readonly string[] = Object.values(AssetType);

const checkIsAssetType = (
	value: null | string,
): value is ValueOf<typeof AssetType> =>
	value !== null && ASSET_TYPES.includes(value);

/**
 * The search lives in the address (`?q=` and `?type=`, research R6), so a
 * refresh, a shared link and Back/Forward show the same search. Submitting
 * adds a history entry; other parameters such as the import panel are kept.
 */
const useAssetSearch = (): AssetSearch => {
	const [searchParameters, setSearchParameters] = useSearchParams();
	const text = searchParameters.get(SearchParameter.QUERY)?.trim() ?? "";
	const type = searchParameters.get(SearchParameter.ASSET_TYPE);

	const activeSearch = useMemo(
		(): AssetSearchRequestDto | null =>
			text === ""
				? null
				: { assetType: checkIsAssetType(type) ? type : null, q: text },
		[text, type],
	);

	const submitSearch = useCallback(
		({ assetType, q }: AssetSearchRequestDto): void => {
			setSearchParameters((current) => {
				const next = new URLSearchParams(current);
				next.set(SearchParameter.QUERY, q);

				if (assetType === null) {
					next.delete(SearchParameter.ASSET_TYPE);
				} else {
					next.set(SearchParameter.ASSET_TYPE, assetType);
				}

				return next;
			});
		},
		[setSearchParameters],
	);

	const clearSearch = useCallback((): void => {
		setSearchParameters((current) => {
			const next = new URLSearchParams(current);
			next.delete(SearchParameter.QUERY);
			next.delete(SearchParameter.ASSET_TYPE);

			return next;
		});
	}, [setSearchParameters]);

	return { activeSearch, clearSearch, submitSearch };
};

export { useAssetSearch };
