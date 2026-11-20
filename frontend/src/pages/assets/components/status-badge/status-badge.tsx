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
	/** Only the coloured shape; the words are given elsewhere (the tree, spec 005 FR-007). */
	isShapeOnly?: boolean;
	status: null | ValueOf<typeof OperationalStatus>;
};

/** A status in words with its shape and colour (spec 004 FR-004). */
const StatusBadge: React.FC<Properties> = ({
	isCompact = false,
	isShapeOnly = false,
	status,
}: Properties) => {
	const tone =
		styles[status === null ? "none" : operationalStatusToTone[status]];

	return isShapeOnly ? (
		<span className={getValidClasses(styles["shape-only"], tone)}>
			<StatusShape status={status} />
		</span>
	) : (
		<span
			className={getValidClasses(
				styles["badge"],
				isCompact && styles["compact"],
				tone,
			)}
		>
			<StatusShape status={status} />
			{getStatusText(status)}
		</span>
	);
};

export { StatusBadge };
