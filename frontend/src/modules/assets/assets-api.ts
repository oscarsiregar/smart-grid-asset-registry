import { APIPath } from "~/libs/enums/enums.js";
import { configureString } from "~/libs/helpers/helpers.js";
import { baseApi } from "~/libs/modules/api/base-api.js";

import { AssetsApiPath, AssetsApiTag } from "./libs/enums/enums.js";
import {
	type AssetDetailDto,
	type AssetListDto,
	type AssetSearchRequestDto,
	type AssetSearchResultDto,
} from "./libs/types/types.js";

const assetsApi = baseApi
	.enhanceEndpoints({ addTagTypes: [AssetsApiTag.ASSET] })
	.injectEndpoints({
		endpoints: (builder) => ({
			getAssetDetail: builder.query<AssetDetailDto, string>({
				// The details area renders every failure, "not found" included.
				extraOptions: { shouldSuppressToast: true },
				providesTags: [AssetsApiTag.ASSET],
				query: (assetId) =>
					configureString(APIPath.ASSETS, AssetsApiPath.$ASSET_ID, {
						assetId: encodeURIComponent(assetId),
					}),
			}),

			getAssetNodes: builder.query<AssetListDto, null | string>({
				// The tree renders its own loading and error rows.
				extraOptions: { shouldSuppressToast: true },
				providesTags: [AssetsApiTag.ASSET],
				query: (parentAssetId) =>
					parentAssetId === null
						? APIPath.ASSETS
						: { params: { parentAssetId }, url: APIPath.ASSETS },
			}),

			searchAssets: builder.query<AssetSearchResultDto, AssetSearchRequestDto>({
				// The results area renders every failure.
				extraOptions: { shouldSuppressToast: true },
				providesTags: [AssetsApiTag.ASSET],
				// No page or limit: the backend's default page of 50 applies, so the
				// frontend keeps no copy of it (principle II).
				query: ({ assetType, q }) => ({
					params: assetType === null ? { q } : { assetType, q },
					url: APIPath.ASSETS,
				}),
			}),
		}),
	});

const { useGetAssetDetailQuery, useGetAssetNodesQuery, useSearchAssetsQuery } =
	assetsApi;

export { useGetAssetDetailQuery, useGetAssetNodesQuery, useSearchAssetsQuery };
