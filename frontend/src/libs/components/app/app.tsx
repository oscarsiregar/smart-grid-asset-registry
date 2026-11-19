import { Outlet as RouterOutlet, ScrollRestoration } from "react-router-dom";

import { OverlayHost } from "~/libs/components/overlay-host/overlay-host.js";

const App: React.FC = () => (
	<OverlayHost>
		<ScrollRestoration />
		<RouterOutlet />
	</OverlayHost>
);

export { App };
