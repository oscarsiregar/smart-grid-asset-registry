import { describe, expect, it } from "vitest";

import { checkIsFailure } from "./check-is-failure.helper.js";

describe("checkIsFailure", () => {
	it.each([
		["no error", undefined, false],
		[
			"a 4xx with an error code",
			{ code: "IMPORT_FILE_EMPTY", message: "x", status: 422 },
			false,
		],
		[
			"a server error",
			{ code: "INTERNAL_SERVER_ERROR", message: "x", status: 500 },
			true,
		],
		["an error without a server shape", { message: "Rejected" }, true],
	])("%s", (_label, error, expected) => {
		expect(checkIsFailure(error)).toBe(expected);
	});
});
