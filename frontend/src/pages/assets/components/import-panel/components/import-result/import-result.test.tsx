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
		render(<ImportResult onReset={vi.fn()} result={VALIDATED} />);

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
		render(<ImportResult onReset={vi.fn()} result={VALIDATED} />);

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

	it("offers to choose another file", () => {
		const handleReset = vi.fn();
		render(<ImportResult onReset={handleReset} result={VALIDATED} />);

		fireEvent.click(
			screen.getByRole("button", { name: "Choose another file" }),
		);

		expect(handleReset).toHaveBeenCalledOnce();
	});
});
