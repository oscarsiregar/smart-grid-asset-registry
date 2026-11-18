import { type ErrorCode } from "~/libs/enums/enums.js";

import { type ServerErrorDetail } from "./server-error-detail.type.js";

type ServerValidationErrorResponse = {
	code: typeof ErrorCode.VALIDATION_FAILED;
	details: ServerErrorDetail[];
	message: string;
};

export { type ServerValidationErrorResponse };
