import { Icon } from "~/libs/components/icon/icon.js";
import { getValidClasses } from "~/libs/helpers/helpers.js";

import { type OutcomeSummary } from "../../libs/types/types.js";
import { outcomeToneToIconName } from "./libs/maps.js";
import styles from "./styles.module.css";

type Properties = OutcomeSummary;

/**
 * States whether data was saved, in words, with an icon and a tone. Colour is
 * never the only signal (constitution, principle V; spec SC-005).
 */
const OutcomeBanner: React.FC<Properties> = ({ message, tone }: Properties) => (
	<p className={getValidClasses(styles["banner"], styles[tone])} role="status">
		<Icon className={styles["icon"]} iconName={outcomeToneToIconName[tone]} />
		<span>{message}</span>
	</p>
);

export { OutcomeBanner };
