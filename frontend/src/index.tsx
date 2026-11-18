import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { Provider as StoreProvider } from "react-redux";
import { Navigate } from "react-router-dom";

import "~/assets/css/styles.css";
import { AppShell } from "~/libs/components/app-shell/app-shell.js";
import { App } from "~/libs/components/app/app.js";
import { Loader } from "~/libs/components/loader/loader.js";
import { RouterProvider } from "~/libs/components/router-provider/router-provider.js";
import { AppRoute } from "~/libs/enums/enums.js";
import { store } from "~/libs/modules/store/store.js";
import { ErrorPage } from "~/pages/error/error.js";
import { NotFound } from "~/pages/not-found/not-found.js";

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
										element: <Navigate replace to={AppRoute.ASSETS} />,
										index: true,
									},
									{
										lazy: async () => {
											const pageModule =
												await import("~/pages/assets/assets.js");

											return { Component: pageModule.Assets };
										},
										path: AppRoute.ASSETS,
									},
									{
										lazy: async () => {
											const pageModule = await import("~/pages/notes/notes.js");

											return { Component: pageModule.Notes };
										},
										path: AppRoute.NOTES,
									},
								],
								element: <AppShell />,
								hydrateFallbackElement: <Loader />,
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
