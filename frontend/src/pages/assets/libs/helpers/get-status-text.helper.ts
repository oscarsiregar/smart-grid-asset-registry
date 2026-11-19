import { type ValueOf } from "~/libs/types/types.js";
import { type OperationalStatus } from "~/modules/assets/assets.js";

import { AssetsPageMessage } from "../enums/enums.js";
import { operationalStatusToLabel } from "../maps.js";

/** An operational status in words; `null` is "No status recorded" (spec FR-009). */
const getStatusText = (
	status: null | ValueOf<typeof OperationalStatus>,
): string =>
	status === null
		? AssetsPageMessage.STATUS_NONE
		: operationalStatusToLabel[status];

export { getStatusText };
