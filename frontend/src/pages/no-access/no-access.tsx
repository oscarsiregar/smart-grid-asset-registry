import noAccessIllustration from "~/assets/img/no-access.svg";
import { FallbackScreen } from "~/libs/components/fallback-screen/fallback-screen.js";
import { AppRoute, HTTPCode } from "~/libs/enums/enums.js";

const noAccessAction = { label: "Back Home", url: AppRoute.DASHBOARD };

const NoAccess: React.FC = () => (
	<main>
		<FallbackScreen
			action={noAccessAction}
			code={HTTPCode.FORBIDDEN}
			illustrationUrl={noAccessIllustration}
			message="You do not have permission to view this page."
			title="Access denied"
		/>
	</main>
);

export { NoAccess };
