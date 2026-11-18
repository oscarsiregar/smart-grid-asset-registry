import { type AppRoute } from "~/libs/enums/enums.js";

import { type ValueOf } from "./value-of.type.js";

type NavigableRoute = Exclude<ValueOf<typeof AppRoute>, typeof AppRoute.ANY>;

export { type NavigableRoute };
