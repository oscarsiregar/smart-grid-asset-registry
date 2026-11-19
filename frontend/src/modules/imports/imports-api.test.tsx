import { act, renderHook, waitFor } from "@testing-library/react";
import { Provider } from "react-redux";
import { afterEach, describe, expect, it, vi } from "vitest";

import { HTTPCode } from "~/libs/enums/enums.js";
import { config } from "~/libs/modules/config/config.js";
import { createStore } from "~/libs/modules/store/store.module.js";
import {
	useGetAssetDetailQuery,
	useGetAssetNodesQuery,
} from "~/modules/assets/assets.js";

import {
	useCreateImportMutation,
	useValidateImportMutation,
} from "./imports-api.js";
import { ImportMode, ImportOutcome, type ImportResultDto } from "./imports.js";

const RESULT: ImportResultDto = {
	counts: { imported: 1, rejected: 0, total: 1, unchanged: 0, warningRows: 0 },
	fileName: "grid.csv",
	import: null,
	isCommitted: false,
	mode: ImportMode.STRICT,
	outcome: ImportOutcome.VALIDATED_ONLY,
	rejectedRows: [],
	warnings: [],
};

afterEach(() => {
	vi.unstubAllGlobals();
});

const createJsonResponse = (status: number, body: unknown): Response =>
	Response.json(body, { status });

const renderValidateMutation = () => {
	const store = createStore(config);
	const wrapper: React.FC<{ children: React.ReactNode }> = ({ children }) => (
		<Provider store={store}>{children}</Provider>
	);

	return renderHook(() => useValidateImportMutation(), { wrapper });
};

describe("validateImport", () => {
	it("posts to /imports/validate and returns the unwrapped result", async () => {
		const fetchMock = vi.fn(() =>
			Promise.resolve(
				createJsonResponse(HTTPCode.OK, {
					code: HTTPCode.OK,
					data: RESULT,
					message: "Successfully validate asset import",
					status: "success",
				}),
			),
		);
		vi.stubGlobal("fetch", fetchMock);
		const file = new File(["asset_id\n"], "grid.csv", { type: "text/csv" });
		const { result } = renderValidateMutation();

		let response: unknown;
		await act(async () => {
			const [validateImport] = result.current;
			response = await validateImport({
				file,
				mode: ImportMode.STRICT,
			}).unwrap();
		});

		expect(response).toEqual(RESULT);
		const [[request]] = fetchMock.mock.calls as unknown as [[Request]];
		expect(request.method).toBe("POST");
		expect(request.url).toBe("http://localhost/api/imports/validate");
		// The multipart body itself is covered by create-import-form-data's
		// test: jsdom's File and Node's fetch cannot read each other's bodies.
	});

	it("exposes a file-level error with its application error code", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn(() =>
				Promise.resolve(
					createJsonResponse(HTTPCode.UNPROCESSED_ENTITY, {
						code: HTTPCode.UNPROCESSED_ENTITY,
						data: { errorCode: "IMPORT_FILE_EMPTY" },
						message: "The file contains no data rows.",
						status: "error",
					}),
				),
			),
		);
		const { result } = renderValidateMutation();

		await act(async () => {
			const [validateImport] = result.current;
			await validateImport({
				file: new File([""], "empty.csv"),
				mode: ImportMode.PARTIAL,
			});
		});

		const [, { error }] = result.current;
		expect(error).toEqual({
			code: "IMPORT_FILE_EMPTY",
			message: "The file contains no data rows.",
			status: HTTPCode.UNPROCESSED_ENTITY,
		});
	});
});

describe("createImport", () => {
	it("posts to /imports and returns the unwrapped result", async () => {
		const committed: ImportResultDto = {
			...RESULT,
			import: { createdAt: "2026-11-19T10:15:00Z", id: "6f1c2c1e" },
			isCommitted: true,
			outcome: ImportOutcome.COMMITTED,
		};
		const fetchMock = vi.fn(() =>
			Promise.resolve(
				Response.json(
					{
						code: HTTPCode.CREATED,
						data: committed,
						message: "Successfully create asset import",
						status: "success",
					},
					{ status: HTTPCode.CREATED },
				),
			),
		);
		vi.stubGlobal("fetch", fetchMock);
		const store = createStore(config);
		const wrapper: React.FC<{ children: React.ReactNode }> = ({ children }) => (
			<Provider store={store}>{children}</Provider>
		);
		const { result } = renderHook(() => useCreateImportMutation(), { wrapper });

		let response: unknown;
		await act(async () => {
			const [createImport] = result.current;
			response = await createImport({
				file: new File(["asset_id"], "grid.csv"),
				mode: ImportMode.PARTIAL,
			}).unwrap();
		});

		expect(response).toEqual(committed);
		const [[request]] = fetchMock.mock.calls as unknown as [[Request]];
		expect(request.method).toBe("POST");
		expect(request.url).toBe("http://localhost/api/imports");
	});
});

// ---- Refreshing the asset explorer after an import (user story 4) ----------

type CreateAnswer = "committed" | "failed" | ImportResultDto;

const ASSET_URLS = [
	"http://localhost/api/assets",
	"http://localhost/api/assets?parentAssetId=SS-001",
	"http://localhost/api/assets/SS-001",
] as const;

const ASSET_DETAIL = {
	asset: {
		assetId: "SS-001",
		assetName: "Aurora",
		assetType: "SUBSTATION",
		commissionedDate: null,
		manufacturer: null,
		model: null,
		operationalStatus: null,
		parentAssetId: null,
		ratingKva: null,
		serialNumber: null,
		voltageKv: null,
	},
	children: [],
	descendantCounts: {
		lvBoards: 0,
		switchboardPanels: 0,
		switchboards: 0,
		transformers: 0,
	},
	path: [],
};

const serveExplorer = (answer: CreateAnswer) => {
	const fetchMock = vi.fn((request: Request): Promise<Response> => {
		if (request.method === "POST") {
			if (answer === "failed") {
				return Promise.resolve(
					createJsonResponse(HTTPCode.INTERNAL_SERVER_ERROR, {
						code: HTTPCode.INTERNAL_SERVER_ERROR,
						data: null,
						message: "Internal Server Error",
						status: "error",
					}),
				);
			}

			const result =
				answer === "committed"
					? {
							...RESULT,
							import: { createdAt: "2026-11-19T10:15:00Z", id: "6f1c2c1e" },
							isCommitted: true,
							outcome: ImportOutcome.COMMITTED,
						}
					: answer;

			return Promise.resolve(
				createJsonResponse(HTTPCode.CREATED, {
					code: HTTPCode.CREATED,
					data: result,
					message: "Successfully create asset import",
					status: "success",
				}),
			);
		}

		const data =
			request.url === "http://localhost/api/assets/SS-001"
				? ASSET_DETAIL
				: { items: [], totalItems: 0 };

		return Promise.resolve(
			createJsonResponse(HTTPCode.OK, {
				code: HTTPCode.OK,
				data,
				message: "ok",
				status: "success",
			}),
		);
	});
	vi.stubGlobal("fetch", fetchMock);

	return fetchMock;
};

const countAssetRequests = (
	fetchMock: ReturnType<typeof serveExplorer>,
): number[] =>
	ASSET_URLS.map(
		(url) =>
			fetchMock.mock.calls.filter(([request]) => request.url === url).length,
	);

// What the explorer has mounted: the top level, an expanded node's children
// and the details area, plus the import panel's mutation.
const useExplorerAndImport = () => ({
	children: useGetAssetNodesQuery("SS-001"),
	createImport: useCreateImportMutation(),
	detail: useGetAssetDetailQuery("SS-001"),
	topLevel: useGetAssetNodesQuery(null),
});

const renderExplorerAndImport = () => {
	const store = createStore(config);
	const wrapper: React.FC<{ children: React.ReactNode }> = ({ children }) => (
		<Provider store={store}>{children}</Provider>
	);

	return renderHook(() => useExplorerAndImport(), { wrapper });
};

const runImport = async (
	result: ReturnType<typeof renderExplorerAndImport>["result"],
): Promise<void> => {
	await act(async () => {
		const [createImport] = result.current.createImport;
		await createImport({
			file: new File(["asset_id"], "grid.csv"),
			mode: ImportMode.PARTIAL,
		});
	});
};

// One request per mounted asset query, or two after a refetch.
const ONE_REQUEST = 1;
const TWO_REQUESTS = 2;
const ONE_REQUEST_EACH = ASSET_URLS.map(() => ONE_REQUEST);
const TWO_REQUESTS_EACH = ASSET_URLS.map(() => TWO_REQUESTS);

describe("createImport and the asset explorer", () => {
	it("refetches every mounted asset query after a committed import", async () => {
		const fetchMock = serveExplorer("committed");
		const { result } = renderExplorerAndImport();
		await waitFor(() => {
			expect(result.current.detail.isSuccess).toBe(true);
		});
		expect(countAssetRequests(fetchMock)).toEqual(ONE_REQUEST_EACH);

		await runImport(result);

		await waitFor(() => {
			expect(countAssetRequests(fetchMock)).toEqual(TWO_REQUESTS_EACH);
		});
	});

	it.each([
		["validated only", RESULT],
		[
			"nothing to commit",
			{ ...RESULT, outcome: ImportOutcome.NOTHING_TO_COMMIT },
		],
		["strict-blocked", { ...RESULT, outcome: ImportOutcome.STRICT_BLOCKED }],
	] as const)(
		"refetches nothing after a result that saved nothing (%s)",
		async (_label, answer) => {
			const fetchMock = serveExplorer(answer);
			const { result } = renderExplorerAndImport();
			await waitFor(() => {
				expect(result.current.detail.isSuccess).toBe(true);
			});

			await runImport(result);

			const [, importState] = result.current.createImport;
			expect(importState.isSuccess).toBe(true);
			expect(countAssetRequests(fetchMock)).toEqual(ONE_REQUEST_EACH);
		},
	);

	it("refetches nothing after a failed import", async () => {
		const fetchMock = serveExplorer("failed");
		const { result } = renderExplorerAndImport();
		await waitFor(() => {
			expect(result.current.detail.isSuccess).toBe(true);
		});

		await runImport(result);

		const [, importState] = result.current.createImport;
		expect(importState.isError).toBe(true);
		expect(countAssetRequests(fetchMock)).toEqual(ONE_REQUEST_EACH);
	});
});
