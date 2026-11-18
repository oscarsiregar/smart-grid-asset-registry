import { type FetchBaseQueryError } from "@reduxjs/toolkit/query/react";

import { HTTPCode } from "~/libs/enums/enums.js";

import {
	PAYLOAD_TOO_LARGE_MESSAGE,
	UNKNOWN_ERROR_MESSAGE,
} from "../constants/constants.js";
import { FetchErrorMessage } from "../enums/enums.js";

/**
 * Message for a failure whose body is not a server error envelope: a network
 * or parsing failure, or an error page from a proxy.
 */
const getFetchErrorMessage = (
	status: FetchBaseQueryError["status"],
): string => {
	if (typeof status === "string") {
		return FetchErrorMessage[status];
	}

	return status === HTTPCode.PAYLOAD_TOO_LARGE
		? PAYLOAD_TOO_LARGE_MESSAGE
		: UNKNOWN_ERROR_MESSAGE;
};

export { getFetchErrorMessage };
