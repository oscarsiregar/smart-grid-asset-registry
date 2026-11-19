import { describe, expect, it } from "vitest";

import { ImportOutcome } from "~/modules/imports/imports.js";

import { getOutcomeSummary } from "./get-outcome-summary.helper.js";

const COUNTS = {
	imported: 205,
	rejected: 17,
	total: 222,
	unchanged: 0,
	warningRows: 2,
};

describe("getOutcomeSummary", () => {
	it.each([
		[ImportOutcome.VALIDATED_ONLY, "Validated only — nothing saved", "info"],
		[ImportOutcome.COMMITTED, "Committed — 205 assets saved", "success"],
		[
			ImportOutcome.NOTHING_TO_COMMIT,
			"Nothing to commit — no new assets",
			"info",
		],
		[
			ImportOutcome.STRICT_BLOCKED,
			"Not committed — strict mode, 17 rows rejected",
			"warning",
		],
	])("states %s in words (FR-027)", (outcome, message, tone) => {
		expect(getOutcomeSummary({ counts: COUNTS, outcome })).toEqual({
			message,
			tone,
		});
	});
});
