/** Asset types as the API sends them (contracts/assets-api.md). */
const AssetType = {
	LV_BOARD: "LV_BOARD",
	SUBSTATION: "SUBSTATION",
	SWITCHBOARD: "SWITCHBOARD",
	SWITCHBOARD_PANEL: "SWITCHBOARD_PANEL",
	TRANSFORMER: "TRANSFORMER",
} as const;

export { AssetType };
