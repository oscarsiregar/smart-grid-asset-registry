import { Navigate, useLocation } from "react-router-dom";

import { Loader } from "~/libs/components/loader/loader.js";
import { Logo } from "~/libs/components/logo/logo.js";
import { AppRoute, ControlSize } from "~/libs/enums/enums.js";
import { getValidClasses } from "~/libs/helpers/helpers.js";
import { useGetAuthenticatedUserQuery } from "~/modules/auth/auth-api.js";

import { SignInForm } from "./components/sign-in-form/sign-in-form.js";
import { SignUpForm } from "./components/sign-up-form/sign-up-form.js";
import { AuthPageMessage } from "./libs/enums/enums.js";
import styles from "./styles.module.css";

const getScreen = (pathname: string): React.JSX.Element => {
	if (pathname === AppRoute.SIGN_UP) {
		return <SignUpForm />;
	}

	return <SignInForm />;
};

const Auth: React.FC = () => {
	const { pathname } = useLocation();
	const { data: user, isLoading } = useGetAuthenticatedUserQuery(undefined);

	if (isLoading) {
		return <Loader />;
	}

	if (user) {
		return <Navigate replace to={AppRoute.DASHBOARD} />;
	}

	return (
		<main className={styles["page"]}>
			<div className={getValidClasses("page-container", styles["container"])}>
				<div className={styles["intro"]}>
					<Logo size={ControlSize.MD} />
					<p className={styles["headline"]}>{AuthPageMessage.HEADLINE}</p>
					<p className={styles["description"]}>{AuthPageMessage.DESCRIPTION}</p>
				</div>
				<div className={styles["card"]}>{getScreen(pathname)}</div>
			</div>
		</main>
	);
};

export { Auth };
