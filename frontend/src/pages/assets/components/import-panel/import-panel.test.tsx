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

const respondWith = (status: number, body: unknown): void => {
	vi.stubGlobal(
		"fetch",
		vi.fn(() => Promise.resolve(Response.json(body, { status }))),
	);
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
		target: { files: [new File(["asset_id\n"], "grid_assets.csv")] },
	});
	fireEvent.click(screen.getByRole("button", { name: "Check file" }));
};

describe("ImportPanel", () => {
	it("shows the validation result: outcome, counts and rejected rows", async () => {
		respondWith(HTTPCode.OK, {
			code: HTTPCode.OK,
			data: VALIDATED,
			message: "Successfully validate asset import",
			status: "success",
		});
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
		respondWith(HTTPCode.UNPROCESSED_ENTITY, {
			code: HTTPCode.UNPROCESSED_ENTITY,
			data: { errorCode: "IMPORT_MISSING_COLUMNS" },
			message: "The file is missing required columns: asset_name.",
			status: "error",
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
		respondWith(HTTPCode.INTERNAL_SERVER_ERROR, {
			code: HTTPCode.INTERNAL_SERVER_ERROR,
			data: null,
			message: "Internal Server Error",
			status: "error",
		});
		renderPanel();

		submitFile();

		expect(
			await screen.findByText(
				"Not checked — the file could not be checked. Nothing was saved.",
			),
		).toBeInTheDocument();
	});

	it("says that nothing was saved when the network fails", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn(() => Promise.reject(new TypeError("Failed to fetch"))),
		);
		renderPanel();

		submitFile();

		expect(
			await screen.findByText(
				"Not checked — the file could not be checked. Nothing was saved.",
			),
		).toBeInTheDocument();
	});

	it("returns to an empty form with Choose another file", async () => {
		respondWith(HTTPCode.OK, {
			code: HTTPCode.OK,
			data: VALIDATED,
			message: "Successfully validate asset import",
			status: "success",
		});
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
