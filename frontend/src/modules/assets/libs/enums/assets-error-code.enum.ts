/** Application error codes the asset endpoints send as `data.errorCode`. */
const AssetsErrorCode = {
	ASSET_NOT_FOUND: "ASSET_NOT_FOUND",
	// A refused read, for example a search text longer than the backend allows.
	INVALID_REQUEST: "INVALID_REQUEST",
} as const;

export { AssetsErrorCode };
