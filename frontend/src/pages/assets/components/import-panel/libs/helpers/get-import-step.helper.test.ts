import { describe, expect, it } from "vitest";

import {
	ImportMode,
	ImportOutcome,
	type ImportResultDto,
} from "~/modules/imports/imports.js";

import { ImportStep } from "../enums/enums.js";
import { getImportStep } from "./get-import-step.helper.js";

const VALIDATED: ImportResultDto = {
	counts: { imported: 1, rejected: 0, total: 1, unchanged: 0, warningRows: 0 },
	fileName: "grid.csv",
	import: null,
	isCommitted: false,
	mode: ImportMode.PARTIAL,
	outcome: ImportOutcome.VALIDATED_ONLY,
	rejectedRows: [],
	warnings: [],
};

describe("getImportStep", () => {
	it("is Choose file without a result", () => {
		expect(getImportStep({ confirmed: undefined, validated: undefined })).toBe(
			ImportStep.CHOOSE_FILE,
		);
	});

	it("is Review after a check", () => {
		expect(getImportStep({ confirmed: undefined, validated: VALIDATED })).toBe(
			ImportStep.REVIEW,
		);
	});

	it.each([
		ImportOutcome.COMMITTED,
		ImportOutcome.NOTHING_TO_COMMIT,
		ImportOutcome.STRICT_BLOCKED,
	])("is Done after a save with outcome %s", (outcome) => {
		expect(
			getImportStep({
				confirmed: { ...VALIDATED, outcome },
				validated: VALIDATED,
			}),
		).toBe(ImportStep.DONE);
	});
});
