import { IconName } from "~/libs/enums/enums.js";
import { type ValueOf } from "~/libs/types/types.js";
import { OperationalStatus } from "~/modules/assets/assets.js";

const operationalStatusToIconName = {
	[OperationalStatus.IN_SERVICE]: IconName.CHECK_CIRCLE,
	[OperationalStatus.MAINTENANCE]: IconName.SETTINGS,
	[OperationalStatus.OUT_OF_SERVICE]: IconName.CIRCLE_X,
} satisfies Record<ValueOf<typeof OperationalStatus>, ValueOf<typeof IconName>>;

export { operationalStatusToIconName };
