import { describe, expect, it } from "vitest";

import { checkIsAddressableAssetId } from "./check-is-addressable-asset-id.helper.js";

describe("checkIsAddressableAssetId", () => {
	it.each([
		["a plain ID", "SS-001", true],
		["an ID with a slash and a space", "SS/9 A", true],
		["an ID that only contains dots", "...", true],
		["a single dot", ".", false],
		["two dots", "..", false],
	])("%s", (_label, assetId, expected) => {
		expect(checkIsAddressableAssetId(assetId)).toBe(expected);
	});
});
