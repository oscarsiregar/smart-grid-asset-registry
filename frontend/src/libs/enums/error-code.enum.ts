/**
 * Error codes the client reacts to. The server sends application error codes
 * as `data.errorCode`; when a response has none, `toServerError` derives one
 * of these from the HTTP status.
 */
const ErrorCode = {
	INTERNAL_SERVER_ERROR: "INTERNAL_SERVER_ERROR",
	NOT_FOUND: "NOT_FOUND",
	PAYLOAD_TOO_LARGE: "PAYLOAD_TOO_LARGE",
	TOO_MANY_REQUESTS: "TOO_MANY_REQUESTS",
	VALIDATION_FAILED: "VALIDATION_FAILED",
} as const;

export { ErrorCode };
