import { ButtonLink } from "~/libs/components/button-link/button-link.js";
import { PageContainer } from "~/libs/components/page-container/page-container.js";
import { PageIntro } from "~/libs/components/page-intro/page-intro.js";
import { AppRoute, IconName } from "~/libs/enums/enums.js";
import { useGetAuthenticatedUserQuery } from "~/modules/auth/auth-api.js";

import { DashboardMessage } from "./libs/enums/enums.js";
import styles from "./styles.module.css";

const Dashboard: React.FC = () => {
	const { data: user } = useGetAuthenticatedUserQuery(undefined);
	const title = user
		? `${DashboardMessage.GREETING}, ${user.nickname}`
		: DashboardMessage.GREETING;

	return (
		<PageContainer>
			<div className={styles["content"]}>
				<PageIntro
					description={DashboardMessage.DESCRIPTION}
					label={DashboardMessage.KICKER}
					title={title}
				/>
				<ButtonLink
					className={styles["action"]}
					iconName={IconName.CLIPBOARD_LIST}
					label={DashboardMessage.NOTES_ACTION}
					to={AppRoute.NOTES}
				/>
			</div>
		</PageContainer>
	);
};

export { Dashboard };
