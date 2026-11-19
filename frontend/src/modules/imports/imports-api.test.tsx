import { act, renderHook } from "@testing-library/react";
import { Provider } from "react-redux";
import { describe, expect, it, vi } from "vitest";

import { HTTPCode } from "~/libs/enums/enums.js";
import { config } from "~/libs/modules/config/config.js";
import { createStore } from "~/libs/modules/store/store.module.js";

import { useValidateImportMutation } from "./imports-api.js";
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
