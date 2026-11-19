import { ErrorCode } from "~/libs/enums/enums.js";
import { checkIsServerError } from "~/libs/modules/api/libs/helpers/check-is-server-error.helper.js";

/**
 * A server or network failure, as opposed to a 4xx answer that carries an
 * error code and a message the operator can act on.
 */
const checkIsFailure = (error: unknown): boolean =>
	error !== undefined &&
	(!checkIsServerError(error) ||
		error.code === ErrorCode.INTERNAL_SERVER_ERROR);

export { checkIsFailure };
