import { useCallback } from "react";
import { Outlet } from "react-router-dom";

import { Button } from "~/libs/components/button/button.js";
import { Link } from "~/libs/components/link/link.js";
import { Logo } from "~/libs/components/logo/logo.js";
import { ButtonVariant, ControlSize, IconName } from "~/libs/enums/enums.js";
import { getValidClasses } from "~/libs/helpers/helpers.js";
import { useGetAuthenticatedUserQuery } from "~/modules/auth/auth-api.js";
import { useSignOut } from "~/modules/auth/auth.js";

import { NAVIGATION_ITEMS } from "./libs/constants/constants.js";
import { ShellLabel } from "./libs/enums/enums.js";
import styles from "./styles.module.css";

const AuthenticatedShell: React.FC = () => {
	const { data: user } = useGetAuthenticatedUserQuery(undefined);
	const signOut = useSignOut();

	const handleSignOutClick = useCallback((): void => {
		void signOut();
	}, [signOut]);

	return (
		<div className={styles["shell"]}>
			<header className={styles["header"]}>
				<div
					className={getValidClasses(
						"page-container",
						styles["header-content"],
					)}
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
					<div className={styles["account"]}>
						{user && (
							<span className={styles["nickname"]}>{user.nickname}</span>
						)}
						<Button
							iconName={IconName.LOG_OUT}
							label={ShellLabel.SIGN_OUT}
							onClick={handleSignOutClick}
							size={ControlSize.SM}
							type="button"
							variant={ButtonVariant.SECONDARY}
						/>
					</div>
				</div>
			</header>
			<main className={styles["main"]}>
				<Outlet />
			</main>
		</div>
	);
};

export { AuthenticatedShell };
