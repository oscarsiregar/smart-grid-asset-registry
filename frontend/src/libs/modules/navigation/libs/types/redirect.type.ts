import { type NavigableRoute } from "~/libs/types/types.js";

type Redirect = {
	shouldReplace: boolean;
	to: NavigableRoute;
};

export { type Redirect };
