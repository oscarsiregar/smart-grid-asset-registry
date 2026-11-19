import { describe, expect, it } from "vitest";

import { AssetType, OperationalStatus } from "~/modules/assets/assets.js";

import { assetTypeToLabel, operationalStatusToLabel } from "./maps.js";

describe("assetTypeToLabel", () => {
	it("names every asset type in words", () => {
		expect(
			Object.values(AssetType).map((type) => assetTypeToLabel[type]),
		).toEqual([
			"LV board",
			"Substation",
			"Switchboard",
			"Switchboard panel",
			"Transformer",
		]);
	});
});

describe("operationalStatusToLabel", () => {
	it("names every status in words", () => {
		expect(
			Object.values(OperationalStatus).map(
				(status) => operationalStatusToLabel[status],
			),
		).toEqual(["In service", "Maintenance", "Out of service"]);
	});
});
