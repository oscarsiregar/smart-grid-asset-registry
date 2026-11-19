import { render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { RejectedRowsTable } from "./rejected-rows-table.js";

const ERROR_LINE_COUNT = 3;

const ROWS = [
	{
		assetId: "SWB-CY-A",
		errors: [
			{
				code: "INVALID_PARENT_TYPE",
				message:
					"A SWITCHBOARD must have a SUBSTATION as parent, but SWB-CY-B is a SWITCHBOARD.",
			},
			{
				code: "CYCLE_DETECTED",
				message: "SWB-CY-A is part of a parent cycle: SWB-CY-A, SWB-CY-B.",
			},
		],
		row: 93,
	},
	{
		assetId: null,
		errors: [{ code: "ASSET_ID_REQUIRED", message: "asset_id is required." }],
		row: 182,
	},
];

describe("RejectedRowsTable", () => {
	it("captions the table with the number of rejected rows", () => {
		render(<RejectedRowsTable rows={ROWS} />);

		expect(
			screen.getByRole("table", { name: "Rejected rows (2)" }),
		).toBeInTheDocument();
	});

	it("shows one line per error with row, asset ID, code and message", () => {
		render(<RejectedRowsTable rows={ROWS} />);

		const [, firstLine, secondLine, ...otherLines] = screen.getAllByRole("row");
		expect([firstLine, secondLine, ...otherLines]).toHaveLength(
			ERROR_LINE_COUNT,
		);
		expect(
			within(firstLine as HTMLElement)
				.getAllByRole("cell")
				.map((cell) => cell.textContent),
		).toEqual([
			"93",
			"SWB-CY-A",
			"INVALID_PARENT_TYPE",
			"A SWITCHBOARD must have a SUBSTATION as parent, but SWB-CY-B is a SWITCHBOARD.",
		]);
		expect(
			within(secondLine as HTMLElement).getByText("CYCLE_DETECTED"),
		).toBeInTheDocument();
	});

	it("marks a missing asset ID visibly and for screen readers", () => {
		render(<RejectedRowsTable rows={ROWS} />);

		const missingIdLine = screen.getAllByRole("row").findLast(Boolean);
		const [, assetIdCell] = within(missingIdLine as HTMLElement).getAllByRole(
			"cell",
		);
		expect(assetIdCell).toHaveTextContent("—none");
	});

	it("says so when no row is rejected", () => {
		render(<RejectedRowsTable rows={[]} />);

		expect(screen.queryByRole("table")).not.toBeInTheDocument();
		expect(screen.getByText("No rows were rejected.")).toBeInTheDocument();
	});

	it("renders an ASSET_ID_CONFLICT message naming the differing fields (FR-019)", () => {
		render(
			<RejectedRowsTable
				rows={[
					{
						assetId: "TX-001-1",
						errors: [
							{
								code: "ASSET_ID_CONFLICT",
								message:
									"Asset TX-001-1 is already stored with different values for: operational_status, rating_kva. Editing stored assets is not supported.",
							},
						],
						row: 3,
					},
				]}
			/>,
		);

		const [, line] = screen.getAllByRole("row");
		expect(
			within(line as HTMLElement)
				.getAllByRole("cell")
				.map((cell) => cell.textContent),
		).toEqual([
			"3",
			"TX-001-1",
			"ASSET_ID_CONFLICT",
			"Asset TX-001-1 is already stored with different values for: operational_status, rating_kva. Editing stored assets is not supported.",
		]);
	});
});
