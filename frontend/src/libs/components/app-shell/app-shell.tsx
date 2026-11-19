import { Outlet } from "react-router-dom";

import styles from "./styles.module.css";

/**
 * The application frame. The only page is the asset page, which owns its
 * sidebar (logo, import action, search and tree) and header, so the frame is
 * just the main landmark.
 */
const AppShell: React.FC = () => (
	<main className={styles["main"]}>
		<Outlet />
	</main>
);

export { AppShell };
