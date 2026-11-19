import { type ValueOf } from "~/libs/types/types.js";
import { OperationalStatus } from "~/modules/assets/assets.js";

// The CSS class (colour) of each status; "none" is used for no status.
const operationalStatusToTone = {
	[OperationalStatus.IN_SERVICE]: "in-service",
	[OperationalStatus.MAINTENANCE]: "maintenance",
	[OperationalStatus.OUT_OF_SERVICE]: "out-of-service",
} satisfies Record<ValueOf<typeof OperationalStatus>, string>;

export { operationalStatusToTone };
