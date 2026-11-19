import { getValidClasses } from "~/libs/helpers/helpers.js";
import { type ValueOf } from "~/libs/types/types.js";
import { type OperationalStatus } from "~/modules/assets/assets.js";

import { getStatusText } from "../../libs/helpers/helpers.js";
import { StatusShape } from "./components/status-shape/status-shape.js";
import { operationalStatusToTone } from "./libs/maps.js";
import styles from "./styles.module.css";

type Properties = {
	/** A smaller, borderless variant for dense rows such as the tree. */
	isCompact?: boolean;
	status: null | ValueOf<typeof OperationalStatus>;
};

/** A status in words with its shape and colour (spec 004 FR-004). */
const StatusBadge: React.FC<Properties> = ({
	isCompact = false,
	status,
}: Properties) => (
	<span
		className={getValidClasses(
			styles["badge"],
			isCompact && styles["compact"],
			styles[status === null ? "none" : operationalStatusToTone[status]],
		)}
	>
		<StatusShape status={status} />
		{getStatusText(status)}
	</span>
);

export { StatusBadge };
