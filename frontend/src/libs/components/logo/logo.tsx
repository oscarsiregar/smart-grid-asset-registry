import { Icon } from "~/libs/components/icon/icon.js";
import { APP_NAME } from "~/libs/constants/constants.js";
import { ControlSize, IconName } from "~/libs/enums/enums.js";
import { getValidClasses } from "~/libs/helpers/helpers.js";
import { type ValueOf } from "~/libs/types/types.js";

import styles from "./styles.module.css";

type Properties = {
	isIconOnly?: boolean;
	size?: ValueOf<typeof ControlSize>;
};

/**
 * The application's mark — a bolt in an accent tile (spec 005 FR-004) — and
 * its name. Alone, the mark is the logo image; beside the name it is
 * decorative.
 */
const Logo: React.FC<Properties> = ({
	isIconOnly = false,
	size = ControlSize.MD,
}: Properties) => (
	<div className={getValidClasses(styles["logo"], styles[size])}>
		{isIconOnly ? (
			<span
				aria-label={`${APP_NAME} logo`}
				className={styles["mark"]}
				data-logo-mark
				role="img"
			>
				<Icon className={styles["mark-icon"]} iconName={IconName.ZAP} />
			</span>
		) : (
			<>
				<span aria-hidden="true" className={styles["mark"]} data-logo-mark>
					<Icon className={styles["mark-icon"]} iconName={IconName.ZAP} />
				</span>
				<span className={styles["wordmark"]}>{APP_NAME}</span>
			</>
		)}
	</div>
);

export { Logo };
