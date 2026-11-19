import { type ValueOf } from "~/libs/types/types.js";
import { OperationalStatus } from "~/modules/assets/assets.js";

import { StatusBadge } from "../../../status-badge/status-badge.js";

type Properties = {
	status: null | ValueOf<typeof OperationalStatus>;
};

/**
 * A status in words with its shape; nothing for an asset in service
 * (feature 002 FR-002).
 */
const StatusLabel: React.FC<Properties> = ({ status }: Properties) =>
	status === OperationalStatus.IN_SERVICE ? null : (
		<StatusBadge isCompact status={status} />
	);

export { StatusLabel };
