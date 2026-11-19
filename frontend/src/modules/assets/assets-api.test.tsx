import { renderHook, waitFor } from "@testing-library/react";
import { Provider } from "react-redux";
import { afterEach, describe, expect, it, vi } from "vitest";

import { HTTPCode } from "~/libs/enums/enums.js";
import { config } from "~/libs/modules/config/config.js";
import { bindShowNotification } from "~/libs/modules/notification/notification.js";
import { createStore } from "~/libs/modules/store/store.module.js";

import {
	type AssetDetailDto,
	type AssetListDto,
	type AssetSearchItemDto,
	type AssetSearchRequestDto,
	type AssetSearchResultDto,
	AssetType,
	OperationalStatus,
	useGetAssetDetailQuery,
	useGetAssetNodesQuery,
	useSearchAssetsQuery,
} from "./assets.js";

const LIST: AssetListDto = {
	items: [
		{
			assetId: "SS-001",
			assetName: "Aurora Substation (North)",
			assetType: AssetType.SUBSTATION,
			childCount: 5,
			operationalStatus: OperationalStatus.IN_SERVICE,
		},
	],
	totalItems: 1,
};

const DETAIL: AssetDetailDto = {
	asset: {
		assetId: "SS/1",
		assetName: "Slash Substation",
		assetType: AssetType.SUBSTATION,
		commissionedDate: "2010-09-11",
		manufacturer: null,
		model: null,
		operationalStatus: null,
		parentAssetId: null,
		ratingKva: null,
		serialNumber: null,
		voltageKv: 22,
	},
	children: [],
	descendantCounts: null,
	path: [
		{
			assetId: "SS/1",
			assetName: "Slash Substation",
			assetType: AssetType.SUBSTATION,
			childCount: 0,
			operationalStatus: null,
		},
	],
};

const mockFetchOnce = (status: number, body: unknown) => {
	const fetchMock = vi.fn(() =>
		Promise.resolve(Response.json(body, { status })),
	);
	vi.stubGlobal("fetch", fetchMock);

	return fetchMock;
};

const success = (data: unknown, message: string) => ({
	code: HTTPCode.OK,
	data,
	message,
	status: "success",
});

const createWrapper = () => {
	const store = createStore(config);
	const wrapper: React.FC<{ children: React.ReactNode }> = ({ children }) => (
		<Provider store={store}>{children}</Provider>
	);

	return wrapper;
};

const getRequest = (fetchMock: ReturnType<typeof vi.fn>): Request => {
	const [[request]] = fetchMock.mock.calls as unknown as [[Request]];

	return request;
};

afterEach(() => {
	vi.unstubAllGlobals();
});

describe("getAssetNodes", () => {
	it("requests the top level from /api/assets and returns the envelope's data", async () => {
		const fetchMock = mockFetchOnce(
			HTTPCode.OK,
			success(LIST, "Successfully get list asset"),
		);

		const { result } = renderHook(() => useGetAssetNodesQuery(null), {
			wrapper: createWrapper(),
		});

		await waitFor(() => {
			expect(result.current.isSuccess).toBe(true);
		});
		expect(result.current.data).toEqual(LIST);
		const request = getRequest(fetchMock);
		expect(request.method).toBe("GET");
		expect(request.url).toBe("http://localhost/api/assets");
	});

	it("encodes the parent ID in the query string", async () => {
		const fetchMock = mockFetchOnce(
			HTTPCode.OK,
			success({ items: [], totalItems: 0 }, "Successfully get list asset"),
		);

		const { result } = renderHook(() => useGetAssetNodesQuery("SS/1"), {
			wrapper: createWrapper(),
		});

		await waitFor(() => {
			expect(result.current.isSuccess).toBe(true);
		});
		expect(getRequest(fetchMock).url).toBe(
			"http://localhost/api/assets?parentAssetId=SS%2F1",
		);
	});
});

describe("getAssetDetail", () => {
	it("encodes the asset ID in the path and returns the envelope's data", async () => {
		const fetchMock = mockFetchOnce(
			HTTPCode.OK,
			success(DETAIL, "Successfully get asset detail"),
		);

		const { result } = renderHook(() => useGetAssetDetailQuery("SS/1"), {
			wrapper: createWrapper(),
		});

		await waitFor(() => {
			expect(result.current.isSuccess).toBe(true);
		});
		expect(result.current.data).toEqual(DETAIL);
		expect(getRequest(fetchMock).url).toBe(
			"http://localhost/api/assets/SS%2F1",
		);
	});

	it("keeps ASSET_NOT_FOUND as the error code and shows no toast", async () => {
		mockFetchOnce(HTTPCode.NOT_FOUND, {
			code: HTTPCode.NOT_FOUND,
			data: { errorCode: "ASSET_NOT_FOUND" },
			message: "No asset with ID NOPE exists.",
			status: "error",
		});
		const showNotificationMock = vi.fn();
		const unbind = bindShowNotification(showNotificationMock);

		const { result } = renderHook(() => useGetAssetDetailQuery("NOPE"), {
			wrapper: createWrapper(),
		});

		await waitFor(() => {
			expect(result.current.isError).toBe(true);
		});
		expect(result.current.error).toEqual({
			code: "ASSET_NOT_FOUND",
			message: "No asset with ID NOPE exists.",
			status: HTTPCode.NOT_FOUND,
		});
		expect(showNotificationMock).not.toHaveBeenCalled();
		unbind();
	});
});

const SEARCH_ITEM: AssetSearchItemDto = {
	ancestorAssetIds: ["SS-001", "SWB-001-1"],
	assetId: "PNL-001-1-01",
	assetName: "Aurora Panel A1",
	assetType: AssetType.SWITCHBOARD_PANEL,
	childCount: 0,
	operationalStatus: OperationalStatus.MAINTENANCE,
};

const SEARCH_RESULT: AssetSearchResultDto = {
	items: [SEARCH_ITEM],
	totalItems: 1,
};

const TEXT_SEARCH: AssetSearchRequestDto = { assetType: null, q: "PNL 1/2" };

describe("searchAssets", () => {
	it("sends the text without page or limit and returns the envelope's data", async () => {
		const fetchMock = mockFetchOnce(
			HTTPCode.OK,
			success(SEARCH_RESULT, "Successfully get list asset"),
		);

		const { result } = renderHook(() => useSearchAssetsQuery(TEXT_SEARCH), {
			wrapper: createWrapper(),
		});

		await waitFor(() => {
			expect(result.current.isSuccess).toBe(true);
		});
		expect(result.current.data).toEqual(SEARCH_RESULT);
		const request = getRequest(fetchMock);
		const url = new URL(request.url);
		expect(request.method).toBe("GET");
		expect(url.pathname).toBe("/api/assets");
		// The text round-trips; no page or limit: the backend's default page of
		// 50 applies (principle II).
		expect(Object.fromEntries(url.searchParams)).toEqual({ q: "PNL 1/2" });
	});

	it("adds the asset type when one is chosen", async () => {
		const fetchMock = mockFetchOnce(
			HTTPCode.OK,
			success(SEARCH_RESULT, "Successfully get list asset"),
		);

		const { result } = renderHook(
			() =>
				useSearchAssetsQuery({
					assetType: AssetType.SWITCHBOARD_PANEL,
					q: "001",
				}),
			{ wrapper: createWrapper() },
		);

		await waitFor(() => {
			expect(result.current.isSuccess).toBe(true);
		});
		const { searchParams } = new URL(getRequest(fetchMock).url);
		expect(Object.fromEntries(searchParams)).toEqual({
			assetType: "SWITCHBOARD_PANEL",
			q: "001",
		});
	});

	it("keeps the error code and shows no toast", async () => {
		mockFetchOnce(HTTPCode.BAD_REQUEST, {
			code: HTTPCode.BAD_REQUEST,
			data: { errorCode: "INVALID_REQUEST" },
			message: "Invalid request: q must be at most 100 characters",
			status: "error",
		});
		const showNotificationMock = vi.fn();
		const unbind = bindShowNotification(showNotificationMock);

		const { result } = renderHook(
			() => useSearchAssetsQuery({ assetType: null, q: "x" }),
			{ wrapper: createWrapper() },
		);

		await waitFor(() => {
			expect(result.current.isError).toBe(true);
		});
		expect(result.current.error).toMatchObject({ code: "INVALID_REQUEST" });
		expect(showNotificationMock).not.toHaveBeenCalled();
		unbind();
	});
});
