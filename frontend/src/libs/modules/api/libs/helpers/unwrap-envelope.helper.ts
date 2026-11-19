const checkIsRecord = (value: unknown): value is Record<string, unknown> => {
	return typeof value === "object" && value !== null;
};

/**
 * The backend answers every success as the go-common envelope
 * `{ code, status: "success", message, data }`. Endpoints work with `data`
 * only, so the envelope is removed here once. Any other body is returned as
 * it is.
 */
const unwrapEnvelope = (body: unknown): unknown => {
	if (
		checkIsRecord(body) &&
		typeof body["code"] === "number" &&
		body["status"] === "success" &&
		"data" in body
	) {
		return body["data"];
	}

	return body;
};

export { unwrapEnvelope };
