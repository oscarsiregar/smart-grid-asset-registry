import { describe, expect, it } from "vitest";

import {
	type AssetDetailDto,
	type AssetDto,
	type AssetNodeDto,
	AssetsApiTag,
	AssetType,
	OperationalStatus,
} from "./assets.js";

// The detail example from contracts/assets-api.md (US1 fields), typed against
// the DTOs so a drift between the contract and the types fails the build.
const ASSET: AssetDto = {
	assetId: "PNL-001-1-3",
	assetName: "Aurora Panel 1-3",
	assetType: AssetType.SWITCHBOARD_PANEL,
	commissionedDate: "2011-06-20",
	manufacturer: "Helios Gridworks",
	model: null,
	operationalStatus: null,
	parentAssetId: "SWB-001-1",
	ratingKva: null,
	serialNumber: "PNL0011C",
	voltageKv: 0.4,
};

const PATH: AssetNodeDto[] = [
	{
		assetId: "SS-001",
		assetName: "Aurora Substation (North)",
		assetType: AssetType.SUBSTATION,
		childCount: 6,
		operationalStatus: OperationalStatus.IN_SERVICE,
	},
	{
		assetId: "SWB-001-1",
		assetName: "Aurora Switchboard 1",
		assetType: AssetType.SWITCHBOARD,
		childCount: 4,
		operationalStatus: OperationalStatus.IN_SERVICE,
	},
	{
		assetId: "PNL-001-1-3",
		assetName: "Aurora Panel 1-3",
		assetType: AssetType.SWITCHBOARD_PANEL,
		childCount: 0,
		operationalStatus: null,
	},
];

const CONTRACT_EXAMPLE: AssetDetailDto = {
	asset: ASSET,
	children: [],
	descendantCounts: null,
	path: PATH,
};

describe("asset DTOs", () => {
	it("match the contract's types and statuses", () => {
		expect(Object.values(AssetType)).toEqual([
			"LV_BOARD",
			"SUBSTATION",
			"SWITCHBOARD",
			"SWITCHBOARD_PANEL",
			"TRANSFORMER",
		]);
		expect(Object.values(OperationalStatus)).toEqual([
			"IN_SERVICE",
			"MAINTENANCE",
			"OUT_OF_SERVICE",
		]);
	});

	it("run the path from the substation down to the asset itself", () => {
		expect(CONTRACT_EXAMPLE.path.map((step) => step.assetId)).toEqual([
			"SS-001",
			ASSET.parentAssetId,
			ASSET.assetId,
		]);
	});

	it("share one cache tag for every asset query", () => {
		expect(Object.values(AssetsApiTag)).toEqual(["Asset"]);
	});
});
