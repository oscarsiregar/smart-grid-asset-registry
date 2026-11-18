import { type ComponentType, StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { Provider as StoreProvider } from "react-redux";
import { Navigate } from "react-router-dom";

import "~/assets/css/styles.css";
import { App } from "~/libs/components/app/app.js";
import { AuthenticatedShell } from "~/libs/components/authenticated-shell/authenticated-shell.js";
import { Loader } from "~/libs/components/loader/loader.js";
import { PrivateRoute } from "~/libs/components/private-route/private-route.js";
import { RouterProvider } from "~/libs/components/router-provider/router-provider.js";
import { AppRoute } from "~/libs/enums/enums.js";
import { store } from "~/libs/modules/store/store.js";
import { ErrorPage } from "~/pages/error/error.js";
import { NotFound } from "~/pages/not-found/not-found.js";

const loadAuthPage = async (): Promise<{ Component: ComponentType }> => {
	const pageModule = await import("~/pages/auth/auth.js");

	return { Component: pageModule.Auth };
};

createRoot(document.querySelector("#root") as HTMLElement).render(
	<StrictMode>
		<StoreProvider store={store}>
			<RouterProvider
				routes={[
					{
						children: [
							{
								children: [
									{
										children: [
											{
												element: <Navigate replace to={AppRoute.DASHBOARD} />,
												index: true,
											},
											{
												lazy: async () => {
													const pageModule =
														await import("~/pages/dashboard/dashboard.js");

													return { Component: pageModule.Dashboard };
												},
												path: AppRoute.DASHBOARD,
											},
											{
												lazy: async () => {
													const pageModule =
														await import("~/pages/notes/notes.js");

													return { Component: pageModule.Notes };
												},
												path: AppRoute.NOTES,
											},
										],
										element: <AuthenticatedShell />,
									},
									{
										lazy: async () => {
											const pageModule =
												await import("~/pages/no-access/no-access.js");

											return { Component: pageModule.NoAccess };
										},
										path: AppRoute.NO_ACCESS,
									},
								],
								element: <PrivateRoute redirectTo={AppRoute.SIGN_IN} />,
								hydrateFallbackElement: <Loader />,
							},
							{
								hydrateFallbackElement: <Loader />,
								lazy: loadAuthPage,
								path: AppRoute.SIGN_IN,
							},
							{
								hydrateFallbackElement: <Loader />,
								lazy: loadAuthPage,
								path: AppRoute.SIGN_UP,
							},
							{
								element: <NotFound />,
								path: AppRoute.ANY,
							},
						],
						element: <App />,
						errorElement: <ErrorPage />,
						path: AppRoute.ROOT,
					},
				]}
			/>
		</StoreProvider>
	</StrictMode>,
);
