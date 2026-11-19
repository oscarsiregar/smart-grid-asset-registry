import { describe, expect, it } from "vitest";

import { getAssetLocationText } from "./get-asset-location-text.helper.js";

describe("getAssetLocationText", () => {
	it("joins the ancestors, substation first", () => {
		expect(getAssetLocationText(["SS-001", "SWB-001-1"])).toBe(
			"SS-001 › SWB-001-1",
		);
	});

	it("names the top level for a substation", () => {
		expect(getAssetLocationText([])).toBe("Top level");
	});
});
