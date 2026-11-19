import { APIPath, HTTPMethod } from "~/libs/enums/enums.js";
import { baseApi } from "~/libs/modules/api/base-api.js";

import { ImportsApiPath } from "./libs/enums/enums.js";
import { createImportFormData } from "./libs/helpers/helpers.js";
import {
	type ImportRequestDto,
	type ImportResultDto,
} from "./libs/types/types.js";

const importsApi = baseApi.injectEndpoints({
	endpoints: (builder) => ({
		validateImport: builder.mutation<ImportResultDto, ImportRequestDto>({
			// The panel renders file-level errors itself.
			extraOptions: { shouldSuppressToast: true },
			query: (payload) => ({
				body: createImportFormData(payload),
				method: HTTPMethod.POST,
				url: `${APIPath.IMPORTS}${ImportsApiPath.VALIDATE}`,
			}),
		}),
	}),
});

const { useValidateImportMutation } = importsApi;

export { useValidateImportMutation };
