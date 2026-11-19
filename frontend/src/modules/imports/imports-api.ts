import { APIPath, HTTPMethod } from "~/libs/enums/enums.js";
import { baseApi } from "~/libs/modules/api/base-api.js";
import { AssetsApiTag } from "~/modules/assets/assets.js";

import { ImportsApiPath } from "./libs/enums/enums.js";
import { createImportFormData } from "./libs/helpers/helpers.js";
import {
	type ImportRequestDto,
	type ImportResultDto,
} from "./libs/types/types.js";

const importsApi = baseApi
	.enhanceEndpoints({ addTagTypes: [AssetsApiTag.ASSET] })
	.injectEndpoints({
		endpoints: (builder) => ({
			createImport: builder.mutation<ImportResultDto, ImportRequestDto>({
				// The panel renders failures itself.
				extraOptions: { shouldSuppressToast: true },
				// Only a commit changes the stored assets; then every mounted asset
				// list and the details refetch (spec FR-026, research R7).
				invalidatesTags: (result) =>
					result?.isCommitted ? [AssetsApiTag.ASSET] : [],
				query: (payload) => ({
					body: createImportFormData(payload),
					method: HTTPMethod.POST,
					url: APIPath.IMPORTS,
				}),
			}),

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

const { useCreateImportMutation, useValidateImportMutation } = importsApi;

export { useCreateImportMutation, useValidateImportMutation };
