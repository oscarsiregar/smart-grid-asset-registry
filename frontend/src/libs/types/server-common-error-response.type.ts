import { type ErrorCode } from "~/libs/enums/enums.js";

import { type ValueOf } from "./value-of.type.js";

type ServerCommonErrorResponse = {
	code: ValueOf<typeof ErrorCode>;
	message: string;
};

export { type ServerCommonErrorResponse };
