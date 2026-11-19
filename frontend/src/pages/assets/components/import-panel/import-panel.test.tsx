import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { Provider } from "react-redux";
import { describe, expect, it, vi } from "vitest";

import { HTTPCode } from "~/libs/enums/enums.js";
import { config } from "~/libs/modules/config/config.js";
import { createStore } from "~/libs/modules/store/store.module.js";
import {
	ImportMode,
	ImportOutcome,
	type ImportResultDto,
} from "~/modules/imports/imports.js";

import { ImportPanel } from "./import-panel.js";

const VALIDATED: ImportResultDto = {
	counts: {
		imported: 205,
		rejected: 17,
		total: 222,
		unchanged: 0,
		warningRows: 2,
	},
	fileName: "grid_assets.csv",
	import: null,
	isCommitted: false,
	mode: ImportMode.PARTIAL,
	outcome: ImportOutcome.VALIDATED_ONLY,
	rejectedRows: [
		{
			assetId: "PNL-M",
			errors: [
				{
					code: "PARENT_NOT_FOUND",
					message: "Parent SWB-404 is neither stored nor in the file.",
				},
			],
			row: 91,
		},
	],
	warnings: [],
};

const COMMITTED: ImportResultDto = {
	...VALIDATED,
	import: { createdAt: "2026-11-19T10:15:00Z", id: "6f1c2c1e" },
	isCommitted: true,
	outcome: ImportOutcome.COMMITTED,
};

const NOTHING_TO_COMMIT: ImportResultDto = {
	...VALIDATED,
	counts: { ...VALIDATED.counts, imported: 0, unchanged: 205 },
	import: { createdAt: "2026-11-19T10:15:00Z", id: "6f1c2c1e" },
	outcome: ImportOutcome.NOTHING_TO_COMMIT,
};

type ServerAnswer = { body: unknown; status: number };

const createSuccess = (
	status: number,
	data: ImportResultDto,
): ServerAnswer => ({
	body: { code: status, data, message: "ok", status: "success" },
	status,
});

const createFailure = (
	status: number,
	message: string,
	errorCode?: string,
): ServerAnswer => ({
	body: {
		code: status,
		data: errorCode === undefined ? null : { errorCode },
		message,
		status: "error",
	},
	status,
});

const INTERNAL_ERROR = createFailure(
	HTTPCode.INTERNAL_SERVER_ERROR,
	"Internal Server Error",
);

/** Answers /imports/validate and /imports separately. */
const serve = (answers: {
	create?: ServerAnswer;
	validate: ServerAnswer;
}): ReturnType<typeof vi.fn> => {
	const fetchMock = vi.fn((request: Request) => {
		const answer = request.url.endsWith("/validate")
			? answers.validate
			: answers.create;

		return answer
			? Promise.resolve(Response.json(answer.body, { status: answer.status }))
			: Promise.reject(new TypeError("Failed to fetch"));
	});
	vi.stubGlobal("fetch", fetchMock);

	return fetchMock;
};

const renderPanel = (): void => {
	render(
		<Provider store={createStore(config)}>
			<ImportPanel />
		</Provider>,
	);
};

const submitFile = (): void => {
	fireEvent.change(screen.getByLabelText("CSV file"), {
		target: { files: [new File(["asset_id"], "grid_assets.csv")] },
	});
	fireEvent.click(screen.getByRole("button", { name: "Check file" }));
};

const NOT_CHECKED_MESSAGE =
	"Not checked — the file could not be checked. Nothing was saved.";

describe("ImportPanel — checking a file (US1)", () => {
	it("shows the validation result: outcome, counts and rejected rows", async () => {
		serve({ validate: createSuccess(HTTPCode.OK, VALIDATED) });
		renderPanel();

		submitFile();

		expect(
			await screen.findByText("Validated only — nothing saved"),
		).toBeInTheDocument();
		expect(
			screen.getByText("Parent SWB-404 is neither stored nor in the file."),
		).toBeInTheDocument();
		expect(screen.queryByLabelText("CSV file")).not.toBeInTheDocument();
	});

	it("shows a file-level problem as an alert with the server message and no table", async () => {
		serve({
			validate: createFailure(
				HTTPCode.UNPROCESSED_ENTITY,
				"The file is missing required columns: asset_name.",
				"IMPORT_MISSING_COLUMNS",
			),
		});
		renderPanel();

		submitFile();

		expect(await screen.findByRole("alert")).toHaveTextContent(
			"The file is missing required columns: asset_name.",
		);
		expect(screen.queryByRole("table")).not.toBeInTheDocument();
		expect(screen.getByLabelText("CSV file")).toBeInTheDocument();
	});

	it("says that nothing was saved when the server fails", async () => {
		serve({ validate: INTERNAL_ERROR });
		renderPanel();

		submitFile();

		expect(await screen.findByText(NOT_CHECKED_MESSAGE)).toBeInTheDocument();
	});

	it("says that nothing was saved when the network fails", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn(() => Promise.reject(new TypeError("Failed to fetch"))),
		);
		renderPanel();

		submitFile();

		expect(await screen.findByText(NOT_CHECKED_MESSAGE)).toBeInTheDocument();
	});

	it("returns to an empty form with Choose another file", async () => {
		serve({ validate: createSuccess(HTTPCode.OK, VALIDATED) });
		renderPanel();
		submitFile();
		await screen.findByText("Validated only — nothing saved");

		fireEvent.click(
			screen.getByRole("button", { name: "Choose another file" }),
		);

		await waitFor(() => {
			expect(screen.getByRole("button", { name: "Check file" })).toBeDisabled();
		});
		expect(
			screen.queryByText("Validated only — nothing saved"),
		).not.toBeInTheDocument();
	});
});

describe("ImportPanel — saving a file (US2)", () => {
	it("saves the same file and shows that data was committed", async () => {
		const fetchMock = serve({
			create: createSuccess(HTTPCode.CREATED, COMMITTED),
			validate: createSuccess(HTTPCode.OK, VALIDATED),
		});
		renderPanel();
		submitFile();
		await screen.findByText("Validated only — nothing saved");

		fireEvent.click(screen.getByRole("button", { name: "Save 205 assets" }));

		expect(
			await screen.findByText("Committed — 205 assets saved"),
		).toBeInTheDocument();
		const urls = fetchMock.mock.calls.map(
			([request]) => (request as Request).url,
		);
		expect(urls).toEqual([
			"http://localhost/api/imports/validate",
			"http://localhost/api/imports",
		]);
		expect(
			screen.queryByRole("button", { name: /Save/ }),
		).not.toBeInTheDocument();
	});

	it("shows Nothing to commit when no new assets were saved", async () => {
		serve({
			create: createSuccess(HTTPCode.CREATED, NOTHING_TO_COMMIT),
			validate: createSuccess(HTTPCode.OK, VALIDATED),
		});
		renderPanel();
		submitFile();
		await screen.findByText("Validated only — nothing saved");

		fireEvent.click(screen.getByRole("button", { name: "Save 205 assets" }));

		expect(
			await screen.findByText("Nothing to commit — no new assets"),
		).toBeInTheDocument();
	});

	it("says the import was rolled back when saving fails, and keeps the result", async () => {
		serve({
			create: INTERNAL_ERROR,
			validate: createSuccess(HTTPCode.OK, VALIDATED),
		});
		renderPanel();
		submitFile();
		await screen.findByText("Validated only — nothing saved");

		fireEvent.click(screen.getByRole("button", { name: "Save 205 assets" }));

		expect(
			await screen.findByText(
				"Not committed — the import failed and was rolled back",
			),
		).toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: "Save 205 assets" }),
		).toBeEnabled();
	});

	it("starts again with Import another file", async () => {
		serve({
			create: createSuccess(HTTPCode.CREATED, COMMITTED),
			validate: createSuccess(HTTPCode.OK, VALIDATED),
		});
		renderPanel();
		submitFile();
		await screen.findByText("Validated only — nothing saved");
		fireEvent.click(screen.getByRole("button", { name: "Save 205 assets" }));
		await screen.findByText("Committed — 205 assets saved");

		fireEvent.click(
			screen.getByRole("button", { name: "Import another file" }),
		);

		await waitFor(() => {
			expect(screen.getByRole("button", { name: "Check file" })).toBeDisabled();
		});
		expect(
			screen.queryByText("Committed — 205 assets saved"),
		).not.toBeInTheDocument();
	});
});
