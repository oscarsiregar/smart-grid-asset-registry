import { type ServerErrorData } from "./server-error-data.type.js";

/**
 * The go-common error envelope: `{ code, status, message, data }`, where
 * `data` is `null` or carries the application error code.
 */
type ServerErrorResponse = {
	code: number;
	data: null | ServerErrorData;
	message: string;
	status: string;
};

export { type ServerErrorResponse };
