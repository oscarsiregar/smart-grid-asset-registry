import { describe, expect, it } from "vitest";

import { getAssetPath } from "./get-asset-path.helper.js";

describe("getAssetPath", () => {
	it.each([
		["a plain ID", "SS-001", "/assets/SS-001"],
		["a slash and a space", "SS/9 A", "/assets/SS%2F9%20A"],
		["a percent sign", "50%", "/assets/50%25"],
		["a question mark and a hash", "x?y#z", "/assets/x%3Fy%23z"],
	])("encodes %s", (_label, assetId, expected) => {
		expect(getAssetPath(assetId)).toBe(expected);
	});
});
