import { type FetchBaseQueryError } from "@reduxjs/toolkit/query/react";

import { ErrorCode, HTTPCode } from "~/libs/enums/enums.js";

/**
 * Derives a client error code for responses that carry no application error
 * code (proxy errors, network failures, unexpected server errors).
 */
const getErrorCodeByStatus = (
	status: FetchBaseQueryError["status"],
): string => {
	switch (status) {
		case HTTPCode.NOT_FOUND: {
			return ErrorCode.NOT_FOUND;
		}

		case HTTPCode.PAYLOAD_TOO_LARGE: {
			return ErrorCode.PAYLOAD_TOO_LARGE;
		}

		case HTTPCode.TOO_MANY_REQUESTS: {
			return ErrorCode.TOO_MANY_REQUESTS;
		}

		default: {
			return ErrorCode.INTERNAL_SERVER_ERROR;
		}
	}
};

export { getErrorCodeByStatus };
