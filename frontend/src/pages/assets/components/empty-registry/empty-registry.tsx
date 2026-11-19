import { AssetsPageMessage } from "../../libs/enums/enums.js";
import styles from "./styles.module.css";

/** No asset is stored yet; the page's Import CSV button opens the panel (spec FR-023). */
const EmptyRegistry: React.FC = () => (
	<p className={styles["empty"]}>{AssetsPageMessage.EMPTY}</p>
);

export { EmptyRegistry };
