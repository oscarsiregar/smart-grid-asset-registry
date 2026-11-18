import { describe, expect, it } from "vitest";

import {
	importFileValidationSchema,
	ImportMode,
	ImportOutcome,
	type ImportRequestDto,
	type ImportResultDto,
} from "./imports.js";

// The ImportResult example from contracts/imports-api.md, typed against the
// DTO so a drift between the contract and the types fails the build.
const CONTRACT_EXAMPLE: ImportResultDto = {
	counts: {
		imported: 205,
		rejected: 17,
		total: 222,
		unchanged: 0,
		warningRows: 2,
	},
	fileName: "grid_assets.csv",
	import: {
		createdAt: "2026-11-19T10:15:00Z",
		id: "6f1c2c1e-6a4e-4b8e-9a52-3d1f2b9c0a11",
	},
	isCommitted: true,
	mode: ImportMode.PARTIAL,
	outcome: ImportOutcome.COMMITTED,
	rejectedRows: [
		{
			assetId: null,
			errors: [{ code: "ASSET_ID_REQUIRED", message: "asset_id is required." }],
			row: 182,
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

const createCsvFile = (content: string, name = "grid_assets.csv"): File =>
	new File([content], name, { type: "text/csv" });

describe("importFileValidationSchema", () => {
	it("accepts a file and a known mode", () => {
		const payload: ImportRequestDto = {
			file: createCsvFile("asset_id\n"),
			mode: ImportMode.STRICT,
		};

		expect(importFileValidationSchema.safeParse(payload).success).toBe(true);
	});

	it("requires a file", () => {
		const result = importFileValidationSchema.safeParse({
			mode: ImportMode.PARTIAL,
		});

		const [firstIssue] = result.error?.issues ?? [];

		expect(result.success).toBe(false);
		expect(firstIssue?.message).toBe("Choose a CSV file.");
	});

	it("rejects an unknown mode", () => {
		const result = importFileValidationSchema.safeParse({
			file: createCsvFile("asset_id\n"),
			mode: "ALL",
		});

		const [firstIssue] = result.error?.issues ?? [];

		expect(result.success).toBe(false);
		expect(firstIssue?.message).toBe("Choose an import mode.");
	});

	it("leaves every rule about the file content to the backend", () => {
		const emptyFile = createCsvFile("", "not-a-csv.txt");

		expect(
			importFileValidationSchema.safeParse({
				file: emptyFile,
				mode: ImportMode.PARTIAL,
			}).success,
		).toBe(true);
	});
});

describe("import DTOs", () => {
	it("match the contract's outcomes and count invariant", () => {
		const { counts } = CONTRACT_EXAMPLE;

		expect(Object.values(ImportOutcome)).toEqual([
			"COMMITTED",
			"NOTHING_TO_COMMIT",
			"STRICT_BLOCKED",
			"VALIDATED_ONLY",
		]);
		expect(counts.imported + counts.unchanged + counts.rejected).toBe(
			counts.total,
		);
	});
});
