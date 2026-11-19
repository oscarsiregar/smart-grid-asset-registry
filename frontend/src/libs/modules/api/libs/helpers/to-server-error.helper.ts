import { type FetchBaseQueryError } from "@reduxjs/toolkit/query/react";

import { ErrorCode } from "~/libs/enums/enums.js";
import {
	type ServerErrorData,
	type ServerErrorResponse,
} from "~/libs/types/types.js";

import { type ServerError } from "../types/server-error.type.js";
import { getErrorCodeByStatus } from "./get-error-code-by-status.helper.js";
import { getFetchErrorMessage } from "./get-fetch-error-message.helper.js";
import { getRetryAfterSeconds } from "./get-retry-after-seconds.helper.js";

const checkIsRecord = (value: unknown): value is Record<string, unknown> => {
	return typeof value === "object" && value !== null;
};

const checkIsServerErrorData = (value: unknown): value is ServerErrorData => {
	return (
		checkIsRecord(value) &&
		typeof value["errorCode"] === "string" &&
		(value["details"] === undefined || Array.isArray(value["details"]))
	);
};

const checkIsServerErrorResponse = (
	payload: unknown,
): payload is ServerErrorResponse => {
	return (
		checkIsRecord(payload) &&
		typeof payload["code"] === "number" &&
		typeof payload["message"] === "string" &&
		(payload["data"] === null || checkIsServerErrorData(payload["data"]))
	);
};

/**
 * Normalizes a failed request into a `ServerError`. The code is the server's
 * `data.errorCode` when the body is a go-common error envelope that carries
 * one; otherwise it is derived from the HTTP status.
 */
const toServerError = (
	error: FetchBaseQueryError,
	headers?: Headers,
): ServerError => {
	if (!checkIsServerErrorResponse(error.data)) {
		return {
			code: getErrorCodeByStatus(error.status),
			message: getFetchErrorMessage(error.status),
			status: error.status,
		};
	}

	const { data, message } = error.data;
	const code = data?.errorCode ?? getErrorCodeByStatus(error.status);

	if (code === ErrorCode.VALIDATION_FAILED && data?.details) {
		return {
			code: ErrorCode.VALIDATION_FAILED,
			details: data.details,
			message,
			status: error.status,
		};
	}

	if (code === ErrorCode.TOO_MANY_REQUESTS) {
		return {
			code: ErrorCode.TOO_MANY_REQUESTS,
			message,
			retryAfterSeconds: getRetryAfterSeconds(headers),
			status: error.status,
		};
	}

	return { code, message, status: error.status };
};

export { toServerError };
