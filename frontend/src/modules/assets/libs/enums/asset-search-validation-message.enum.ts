/**
 * Messages for the search form. It only checks that there is text to search;
 * the length limit and matching are the backend's (principle II).
 */
const AssetSearchValidationMessage = {
	TEXT_REQUIRED: "Enter an asset ID or name to search.",
} as const;

export { AssetSearchValidationMessage };
