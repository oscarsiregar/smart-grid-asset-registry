import { type FetchBaseQueryError } from "@reduxjs/toolkit/query/react";

import { type ErrorCode } from "~/libs/enums/enums.js";

type ServerRateLimitError = {
	code: typeof ErrorCode.TOO_MANY_REQUESTS;
	message: string;
	retryAfterSeconds: null | number;
	status: FetchBaseQueryError["status"];
};

export { type ServerRateLimitError };
