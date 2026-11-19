import { type ServerErrorDetail } from "./server-error-detail.type.js";

/**
 * The `data` of an error response that carries an application error code
 * (constitution, Template Exception 4). `details` is only present for
 * field-level validation errors.
 */
type ServerErrorData = {
	details?: ServerErrorDetail[];
	errorCode: string;
};

export { type ServerErrorData };
