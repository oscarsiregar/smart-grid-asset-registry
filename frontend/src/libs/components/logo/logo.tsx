import logo from "~/assets/img/logo.svg";
import { APP_NAME } from "~/libs/constants/constants.js";
import { ControlSize } from "~/libs/enums/enums.js";
import { getValidClasses } from "~/libs/helpers/helpers.js";
import { type ValueOf } from "~/libs/types/types.js";

import styles from "./styles.module.css";

type Properties = {
	isIconOnly?: boolean;
	size?: ValueOf<typeof ControlSize>;
};

const Logo: React.FC<Properties> = ({
	isIconOnly = false,
	size = ControlSize.MD,
}: Properties) => (
	<div className={getValidClasses(styles["logo"], styles[size])}>
		<img
			alt={isIconOnly ? `${APP_NAME} logo` : ""}
			className={styles["mark"]}
			src={logo}
		/>
		{!isIconOnly && <span className={styles["wordmark"]}>{APP_NAME}</span>}
	</div>
);

export { Logo };
