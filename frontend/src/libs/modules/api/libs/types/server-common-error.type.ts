import { type FetchBaseQueryError } from "@reduxjs/toolkit/query/react";

import { type ErrorCode } from "~/libs/enums/enums.js";
import { type ValueOf } from "~/libs/types/types.js";

type ServerCommonError = {
	code: ValueOf<typeof ErrorCode>;
	message: string;
	status: FetchBaseQueryError["status"];
};

export { type ServerCommonError };
