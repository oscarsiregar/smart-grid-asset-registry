import { Navigate, Outlet } from "react-router-dom";

import { Loader } from "~/libs/components/loader/loader.js";
import { type NavigableRoute } from "~/libs/types/types.js";
import { useGetAuthenticatedUserQuery } from "~/modules/auth/auth-api.js";

type Properties = {
	redirectTo: NavigableRoute;
};

const PrivateRoute: React.FC<Properties> = ({ redirectTo }: Properties) => {
	const { data: user, isLoading } = useGetAuthenticatedUserQuery(undefined);

	if (isLoading) {
		return <Loader />;
	}

	if (!user) {
		return <Navigate to={redirectTo} />;
	}

	return <Outlet />;
};

export { PrivateRoute };
