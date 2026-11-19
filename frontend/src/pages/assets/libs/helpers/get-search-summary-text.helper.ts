import { EMPTY_LENGTH } from "~/libs/constants/constants.js";
import { configureString } from "~/libs/helpers/helpers.js";
import { type AssetSearchRequestDto } from "~/modules/assets/assets.js";

import { SINGLE_MATCH_COUNT } from "../constants/constants.js";
import { AssetsPageMessage } from "../enums/enums.js";
import { assetTypeToLabel } from "../maps.js";

type Options = {
	receivedCount: number;
	search: AssetSearchRequestDto;
	totalItems: number;
};

// " of type <Type>" when the search has one.
const getTypeSuffix = ({ assetType }: AssetSearchRequestDto): string =>
	assetType === null
		? ""
		: configureString(AssetsPageMessage.RESULTS_TYPE_SUFFIX, {
				type: assetTypeToLabel[assetType],
			});

/**
 * The search summary: no match, more matches than received (the count shown
 * is the number of items received, never a frontend limit), or the count.
 */
const getSearchSummaryText = ({
	receivedCount,
	search,
	totalItems,
}: Options): string => {
	const text = search.q;

	if (totalItems === EMPTY_LENGTH) {
		return (
			configureString(AssetsPageMessage.NO_MATCH, { text }) +
			getTypeSuffix(search)
		);
	}

	if (receivedCount < totalItems) {
		return configureString(AssetsPageMessage.RESULTS_PARTIAL, {
			count: String(totalItems),
			shown: String(receivedCount),
			text,
		});
	}

	const matches = configureString(
		totalItems === SINGLE_MATCH_COUNT
			? AssetsPageMessage.RESULTS_SUMMARY_ONE
			: AssetsPageMessage.RESULTS_SUMMARY,
		{ count: String(totalItems), text },
	);

	return matches + getTypeSuffix(search);
};

export { getSearchSummaryText };
