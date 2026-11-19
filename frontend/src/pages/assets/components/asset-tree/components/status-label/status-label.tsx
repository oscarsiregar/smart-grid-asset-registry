import { Icon } from "~/libs/components/icon/icon.js";
import { IconName } from "~/libs/enums/enums.js";
import { type ValueOf } from "~/libs/types/types.js";
import { OperationalStatus } from "~/modules/assets/assets.js";

import { AssetsPageMessage } from "../../../../libs/enums/enums.js";
import { operationalStatusToLabel } from "../../../../libs/maps.js";
import { operationalStatusToIconName } from "./libs/maps.js";
import styles from "./styles.module.css";

type Properties = {
	status: null | ValueOf<typeof OperationalStatus>;
};

/** A status in words with an icon; nothing for an asset in service (spec FR-002). */
const StatusLabel: React.FC<Properties> = ({ status }: Properties) => {
	if (status === OperationalStatus.IN_SERVICE) {
		return null;
	}

	const label =
		status === null
			? AssetsPageMessage.STATUS_NONE
			: operationalStatusToLabel[status];
	const iconName =
		status === null ? IconName.INFO : operationalStatusToIconName[status];

	return (
		<span className={styles["label"]}>
			<Icon className={styles["icon"]} iconName={iconName} />
			{label}
		</span>
	);
};

export { StatusLabel };
