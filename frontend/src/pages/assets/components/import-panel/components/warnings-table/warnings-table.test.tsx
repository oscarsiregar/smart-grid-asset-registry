import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { WarningsTable } from "./warnings-table.js";

const HEADER_AND_TWO_WARNINGS = 3;

const WARNINGS = [
	{
		assetId: "TX-D1",
		code: "DUPLICATE_SERIAL_NUMBER",
		message: "Serial number SER-TX-DUP is also used by TX-D2.",
		row: 89,
	},
	{
		assetId: "TX-D2",
		code: "DUPLICATE_SERIAL_NUMBER",
		message: "Serial number SER-TX-DUP is also used by TX-D1.",
		row: 90,
	},
];

describe("WarningsTable", () => {
	it("lists every warning with row, asset ID, code and message", () => {
		render(<WarningsTable warnings={WARNINGS} />);

		expect(
			screen.getByRole("table", { name: "Warnings (2)" }),
		).toBeInTheDocument();
		expect(screen.getAllByRole("row")).toHaveLength(HEADER_AND_TWO_WARNINGS);
		expect(
			screen.getByText("Serial number SER-TX-DUP is also used by TX-D2."),
		).toBeInTheDocument();
	});

	it("says so when there are no warnings", () => {
		render(<WarningsTable warnings={[]} />);

		expect(screen.queryByRole("table")).not.toBeInTheDocument();
		expect(screen.getByText("No rows have warnings.")).toBeInTheDocument();
	});
});
