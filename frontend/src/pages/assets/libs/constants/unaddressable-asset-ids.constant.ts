/**
 * Asset IDs that are path dot segments: browsers and `fetch` rewrite them, so
 * they have no address and no detail request (spec edge case "IDs . and ..").
 */
const UNADDRESSABLE_ASSET_IDS: readonly string[] = [".", ".."];

export { UNADDRESSABLE_ASSET_IDS };
