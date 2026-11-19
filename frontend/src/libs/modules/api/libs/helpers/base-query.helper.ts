import {
	type BaseQueryFn,
	type FetchArgs,
	fetchBaseQuery,
} from "@reduxjs/toolkit/query/react";

import { NotificationType } from "~/libs/components/overlay-host/libs/enums/enums.js";
import { ErrorCode } from "~/libs/enums/enums.js";
import { config } from "~/libs/modules/config/config.js";
import { showNotification } from "~/libs/modules/notification/notification.js";

import { type BaseQueryExtraOptions } from "../types/base-query-extra-options.type.js";
import { type ServerError } from "../types/server-error.type.js";
import { toServerError } from "./to-server-error.helper.js";
import { unwrapEnvelope } from "./unwrap-envelope.helper.js";

type BaseQueryFunctionInternal = BaseQueryFn<
	FetchArgs | string,
	unknown,
	ServerError,
	BaseQueryExtraOptions | undefined
>;

// The app has no users (constitution, Template Exception 1): requests
// carry no token and there is no 401/403 handling.
const fetchQuery = fetchBaseQuery({
	baseUrl: config.ENV.API.ORIGIN_URL,
});

const baseQuery: BaseQueryFunctionInternal = async (
	arguments_,
	api,
	extraOptions,
) => {
	const result = await fetchQuery(arguments_, api, extraOptions ?? {});

	if (!result.error) {
		return { ...result, data: unwrapEnvelope(result.data) };
	}

	const error = toServerError(result.error, result.meta?.response?.headers);

	switch (error.code) {
		case ErrorCode.INTERNAL_SERVER_ERROR: {
			showNotification({
				id: error.code,
				message: error.message,
				type: NotificationType.DANGER,
			});

			return { error };
		}

		case ErrorCode.VALIDATION_FAILED: {
			return { error };
		}

		default: {
			if (!extraOptions?.shouldSuppressToast) {
				showNotification({
					id: error.code,
					message: error.message,
					type: NotificationType.DANGER,
				});
			}

			return { error };
		}
	}
};

export { baseQuery };
