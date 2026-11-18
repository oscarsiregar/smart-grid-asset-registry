import { Outlet } from "react-router-dom";

import { Link } from "~/libs/components/link/link.js";
import { Logo } from "~/libs/components/logo/logo.js";
import { ControlSize } from "~/libs/enums/enums.js";
import { getValidClasses } from "~/libs/helpers/helpers.js";

import { NAVIGATION_ITEMS } from "./libs/constants/constants.js";
import { ShellLabel } from "./libs/enums/enums.js";
import styles from "./styles.module.css";

const AppShell: React.FC = () => (
	<div className={styles["shell"]}>
		<header className={styles["header"]}>
			<div
				className={getValidClasses("page-container", styles["header-content"])}
			>
				<Logo size={ControlSize.SM} />
				<nav aria-label={ShellLabel.NAVIGATION}>
					<ul className={styles["navigation-list"]}>
						{NAVIGATION_ITEMS.map(({ label, to }) => (
							<li key={to}>
								<Link className={styles["navigation-link"]} to={to}>
									{label}
								</Link>
							</li>
						))}
					</ul>
				</nav>
			</div>
		</header>
		<main className={styles["main"]}>
			<Outlet />
		</main>
	</div>
);

export { AppShell };
