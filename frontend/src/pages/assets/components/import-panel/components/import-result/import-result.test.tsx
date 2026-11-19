import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import {
	ImportMode,
	ImportOutcome,
	type ImportResultDto,
} from "~/modules/imports/imports.js";

import { ImportResult } from "./import-result.js";

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
			assetId: "TX-BD",
			errors: [
				{
					code: "INVALID_DATE",
					message: "commissioned_date 2026-13-41 is not a valid date.",
				},
			],
			row: 179,
		},
	],
	warnings: [
		{
			assetId: "TX-D1",
			code: "DUPLICATE_SERIAL_NUMBER",
			message: "Serial number SER-TX-DUP is also used by TX-D2.",
			row: 89,
		},
	],
};

const getCount = (label: string): null | string => {
	const term = within(screen.getByLabelText("Summary")).getByText(label);

	return term.nextElementSibling?.textContent ?? null;
};

describe("ImportResult", () => {
	it("states that nothing was saved and shows the counts with 'to import'", () => {
		render(
			<ImportResult
				onReset={vi.fn()}
				resetLabel="Choose another file"
				result={VALIDATED}
			/>,
		);

		expect(screen.getByRole("status")).toHaveTextContent(
			"Validated only — nothing saved",
		);
		expect(getCount("Rows in file")).toBe("222");
		expect(getCount("To import")).toBe("205");
		expect(getCount("Unchanged")).toBe("0");
		expect(getCount("Rejected")).toBe("17");
		expect(getCount("Rows with warnings")).toBe("2");
	});

	it("lists the rejected rows and warnings", () => {
		render(
			<ImportResult
				onReset={vi.fn()}
				resetLabel="Choose another file"
				result={VALIDATED}
			/>,
		);

		expect(
			screen.getByRole("table", { name: "Rejected rows (1)" }),
		).toBeInTheDocument();
		expect(
			screen.getByRole("table", { name: "Warnings (1)" }),
		).toBeInTheDocument();
		expect(screen.getByText("INVALID_DATE")).toBeInTheDocument();
	});

	it("labels the count 'Imported' once data is committed", () => {
		render(
			<ImportResult
				onReset={vi.fn()}
				resetLabel="Import another file"
				result={{
					...VALIDATED,
					isCommitted: true,
					outcome: ImportOutcome.COMMITTED,
				}}
			/>,
		);

		expect(getCount("Imported")).toBe("205");
		expect(screen.queryByText("To import")).not.toBeInTheDocument();
	});

	it("keeps listing every rejected row when strict mode blocked the commit", () => {
		render(
			<ImportResult
				onReset={vi.fn()}
				resetLabel="Import another file"
				result={{
					...VALIDATED,
					mode: ImportMode.STRICT,
					outcome: ImportOutcome.STRICT_BLOCKED,
				}}
			/>,
		);

		expect(screen.getByRole("status")).toHaveTextContent(
			"Not committed — strict mode, 17 rows rejected",
		);
		expect(
			screen.getByRole("table", { name: "Rejected rows (1)" }),
		).toBeInTheDocument();
		expect(screen.getByText("INVALID_DATE")).toBeInTheDocument();
		expect(getCount("To import")).toBe("205");
		expect(
			screen.queryByRole("button", { name: /Save/ }),
		).not.toBeInTheDocument();
	});

	it("shows a re-import: nothing to commit, every row unchanged and the same rejections", () => {
		render(
			<ImportResult
				onReset={vi.fn()}
				resetLabel="Import another file"
				result={{
					...VALIDATED,
					counts: { ...VALIDATED.counts, imported: 0, unchanged: 205 },
					import: { createdAt: "2026-11-19T10:15:00Z", id: "6f1c2c1e" },
					outcome: ImportOutcome.NOTHING_TO_COMMIT,
				}}
			/>,
		);

		expect(screen.getByRole("status")).toHaveTextContent(
			"Nothing to commit — no new assets",
		);
		expect(getCount("Unchanged")).toBe("205");
		expect(getCount("To import")).toBe("0");
		expect(getCount("Rejected")).toBe("17");
		expect(
			screen.getByRole("table", { name: "Rejected rows (1)" }),
		).toBeInTheDocument();
	});

	it("offers to choose another file", () => {
		const handleReset = vi.fn();
		render(
			<ImportResult
				onReset={handleReset}
				resetLabel="Choose another file"
				result={VALIDATED}
			/>,
		);

		fireEvent.click(
			screen.getByRole("button", { name: "Choose another file" }),
		);

		expect(handleReset).toHaveBeenCalledOnce();
	});

	it("offers Save with the number of new assets when confirming is possible", () => {
		const handleConfirm = vi.fn();
		render(
			<ImportResult
				onConfirm={handleConfirm}
				onReset={vi.fn()}
				resetLabel="Choose another file"
				result={VALIDATED}
			/>,
		);

		fireEvent.click(screen.getByRole("button", { name: "Save 205 assets" }));

		expect(handleConfirm).toHaveBeenCalledOnce();
	});

	it("shows that the import is being saved and blocks both actions", () => {
		render(
			<ImportResult
				isSaving
				onConfirm={vi.fn()}
				onReset={vi.fn()}
				resetLabel="Choose another file"
				result={VALIDATED}
			/>,
		);

		expect(screen.getByText("Saving…")).toHaveAttribute("role", "status");
		expect(
			screen.getByRole("button", { name: /Save 205 assets/ }),
		).toBeDisabled();
		expect(
			screen.getByRole("button", { name: "Choose another file" }),
		).toBeDisabled();
	});

	it("offers no Save without onConfirm", () => {
		render(
			<ImportResult
				onReset={vi.fn()}
				resetLabel="Import another file"
				result={{
					...VALIDATED,
					isCommitted: true,
					outcome: ImportOutcome.COMMITTED,
				}}
			/>,
		);

		expect(screen.getByRole("status")).toHaveTextContent(
			"Committed — 205 assets saved",
		);
		expect(
			screen.queryByRole("button", { name: /Save/ }),
		).not.toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: "Import another file" }),
		).toBeEnabled();
	});
});
