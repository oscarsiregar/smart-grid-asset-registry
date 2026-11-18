import { type ServerCommonErrorResponse } from "./server-common-error-response.type.js";
import { type ServerValidationErrorResponse } from "./server-validation-error-response.type.js";

type ServerErrorResponse =
	ServerCommonErrorResponse | ServerValidationErrorResponse;

export { type ServerErrorResponse };
