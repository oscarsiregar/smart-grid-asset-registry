import { type FetchBaseQueryError } from "@reduxjs/toolkit/query/react";

import { type ErrorCode } from "~/libs/enums/enums.js";
import { type ServerErrorDetail } from "~/libs/types/types.js";

type ServerValidationError = {
	code: typeof ErrorCode.VALIDATION_FAILED;
	details: ServerErrorDetail[];
	message: string;
	status: FetchBaseQueryError["status"];
};

export { type ServerValidationError };
