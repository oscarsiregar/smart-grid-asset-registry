import { AssetsPageMessage } from "../../../../libs/enums/enums.js";
import styles from "./styles.module.css";

/**
 * The column headings over the children rows (spec 005 FR-014). Decorative:
 * each row already reads its name, ID and status (research R8).
 */
const ChildColumns: React.FC = () => (
	<div
		aria-hidden="true"
		className={styles["columns"]}
		data-testid="child-columns"
	>
		<span>{AssetsPageMessage.COLUMN_NAME}</span>
		<span>{AssetsPageMessage.COLUMN_ASSET_ID}</span>
		<span>{AssetsPageMessage.COLUMN_STATUS}</span>
	</div>
);

export { ChildColumns };
