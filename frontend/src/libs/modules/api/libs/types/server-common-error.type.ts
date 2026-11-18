import { type FetchBaseQueryError } from "@reduxjs/toolkit/query/react";

type ServerCommonError = {
	/** An `ErrorCode` value or an application error code from the server. */
	code: string;
	message: string;
	status: FetchBaseQueryError["status"];
};

export { type ServerCommonError };
