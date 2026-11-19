import { generatePath } from "react-router-dom";

import {
	type GeneratedRoutePath,
	type NavigableRoute,
} from "~/libs/types/types.js";

/**
 * Fills the route's parameters (each `:name`), percent-encoding every value,
 * and marks the result as a generated address that `Link` accepts. Takes
 * exactly `generatePath`'s arguments, limited to the app's routes.
 */
const generateRoutePath = <Route extends NavigableRoute>(
	...arguments_: Parameters<typeof generatePath<Route>>
): GeneratedRoutePath =>
	// The only place the brand is applied: the value really is a filled route.
	generatePath(...arguments_) as GeneratedRoutePath;

export { generateRoutePath };
