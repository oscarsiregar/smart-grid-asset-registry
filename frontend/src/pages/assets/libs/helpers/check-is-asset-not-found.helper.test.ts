import { describe, expect, it } from "vitest";

import { checkIsAssetNotFound } from "./check-is-asset-not-found.helper.js";

describe("checkIsAssetNotFound", () => {
	it.each([
		[
			"ASSET_NOT_FOUND",
			{ code: "ASSET_NOT_FOUND", message: "x", status: 404 },
			true,
		],
		[
			"a 404 without the code",
			{ code: "NOT_FOUND", message: "x", status: 404 },
			false,
		],
		[
			"a server error",
			{ code: "INTERNAL_SERVER_ERROR", message: "x", status: 500 },
			false,
		],
		["no error", undefined, false],
		["an error without a server shape", new Error("x"), false],
	])("%s", (_label, error, expected) => {
		expect(checkIsAssetNotFound(error)).toBe(expected);
	});
});
