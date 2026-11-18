const AppRoute = {
	ANY: "*",
	DASHBOARD: "/dashboard",
	NO_ACCESS: "/no-access",
	NOTES: "/notes",
	ROOT: "/",
	SIGN_IN: "/sign-in",
	SIGN_UP: "/sign-up",
} as const;

export { AppRoute };
