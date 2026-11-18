import { type ServerCommonError } from "./server-common-error.type.js";
import { type ServerRateLimitError } from "./server-rate-limit-error.type.js";
import { type ServerValidationError } from "./server-validation-error.type.js";

type ServerError =
	ServerCommonError | ServerRateLimitError | ServerValidationError;

export { type ServerError };
