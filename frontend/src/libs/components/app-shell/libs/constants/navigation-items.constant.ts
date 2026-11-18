import { AppRoute } from "~/libs/enums/enums.js";

import { type NavigationItem } from "../types/types.js";

const NAVIGATION_ITEMS: readonly NavigationItem[] = [
	{ label: "Notes", to: AppRoute.NOTES },
];

export { NAVIGATION_ITEMS };
