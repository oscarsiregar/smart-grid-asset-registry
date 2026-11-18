import { type FetchBaseQueryError } from "@reduxjs/toolkit/query/react";
import { describe, expect, it } from "vitest";

import { ErrorCode, HTTPCode } from "~/libs/enums/enums.js";

import { PAYLOAD_TOO_LARGE_MESSAGE } from "../constants/constants.js";
import { FetchErrorMessage } from "../enums/enums.js";
import { toServerError } from "./to-server-error.helper.js";

const RETRY_AFTER_SECONDS = 30;

const createEnvelope = (
	status: number,
	message: string,
	data: unknown,
): FetchBaseQueryError => ({
	data: { code: status, data, message, status: "error" },
	status,
});

describe("toServerError", () => {
	it("uses data.errorCode from the go-common envelope", () => {
		const error = createEnvelope(
			HTTPCode.UNPROCESSED_ENTITY,
			"The file has no data rows",
			{ errorCode: "IMPORT_FILE_EMPTY" },
		);

		expect(toServerError(error)).toEqual({
			code: "IMPORT_FILE_EMPTY",
			message: "The file has no data rows",
			status: HTTPCode.UNPROCESSED_ENTITY,
		});
	});

	it.each([
		[HTTPCode.INTERNAL_SERVER_ERROR, ErrorCode.INTERNAL_SERVER_ERROR],
		[HTTPCode.PAYLOAD_TOO_LARGE, ErrorCode.PAYLOAD_TOO_LARGE],
		[HTTPCode.NOT_FOUND, ErrorCode.NOT_FOUND],
	])(
		"derives the code from status %i when data is null",
		(status, expectedCode) => {
			const error = createEnvelope(status, "Server message", null);

			expect(toServerError(error)).toEqual({
				code: expectedCode,
				message: "Server message",
				status,
			});
		},
	);

	it("adds retryAfterSeconds for 429", () => {
		const error = createEnvelope(
			HTTPCode.TOO_MANY_REQUESTS,
			"Too many requests, please try again later",
			null,
		);
		const headers = new Headers({ "retry-after": String(RETRY_AFTER_SECONDS) });

		expect(toServerError(error, headers)).toEqual({
			code: ErrorCode.TOO_MANY_REQUESTS,
			message: "Too many requests, please try again later",
			retryAfterSeconds: RETRY_AFTER_SECONDS,
			status: HTTPCode.TOO_MANY_REQUESTS,
		});
	});

	it("keeps field details for validation errors", () => {
		const details = [{ message: "Too long", path: ["title"] }];
		const error = createEnvelope(HTTPCode.UNPROCESSED_ENTITY, "Invalid", {
			details,
			errorCode: ErrorCode.VALIDATION_FAILED,
		});

		expect(toServerError(error)).toEqual({
			code: ErrorCode.VALIDATION_FAILED,
			details,
			message: "Invalid",
			status: HTTPCode.UNPROCESSED_ENTITY,
		});
	});

	it("handles a proxy 413 page that is not an envelope", () => {
		const error: FetchBaseQueryError = {
			data: "<html>413 Request Entity Too Large</html>",
			status: HTTPCode.PAYLOAD_TOO_LARGE,
		};

		expect(toServerError(error)).toEqual({
			code: ErrorCode.PAYLOAD_TOO_LARGE,
			message: PAYLOAD_TOO_LARGE_MESSAGE,
			status: HTTPCode.PAYLOAD_TOO_LARGE,
		});
	});

	it("handles a network failure", () => {
		const error: FetchBaseQueryError = {
			error: "TypeError: Failed to fetch",
			status: "FETCH_ERROR",
		};

		expect(toServerError(error)).toEqual({
			code: ErrorCode.INTERNAL_SERVER_ERROR,
			message: FetchErrorMessage.FETCH_ERROR,
			status: "FETCH_ERROR",
		});
	});
});
