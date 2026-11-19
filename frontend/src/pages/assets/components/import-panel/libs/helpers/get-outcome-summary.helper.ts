import { configureString } from "~/libs/helpers/helpers.js";
import {
	ImportOutcome,
	type ImportResultDto,
} from "~/modules/imports/imports.js";

import { ImportOutcomeMessage, OutcomeTone } from "../enums/enums.js";
import { type OutcomeSummary } from "../types/types.js";

/** The outcome text and tone of a server result (spec FR-027). */
const getOutcomeSummary = ({
	counts,
	outcome,
}: Pick<ImportResultDto, "counts" | "outcome">): OutcomeSummary => {
	switch (outcome) {
		case ImportOutcome.COMMITTED: {
			return {
				message: configureString(ImportOutcomeMessage.COMMITTED, {
					count: String(counts.imported),
				}),
				tone: OutcomeTone.SUCCESS,
			};
		}

		case ImportOutcome.NOTHING_TO_COMMIT: {
			return {
				message: ImportOutcomeMessage.NOTHING_TO_COMMIT,
				tone: OutcomeTone.INFO,
			};
		}

		case ImportOutcome.STRICT_BLOCKED: {
			return {
				message: configureString(ImportOutcomeMessage.STRICT_BLOCKED, {
					count: String(counts.rejected),
				}),
				tone: OutcomeTone.WARNING,
			};
		}

		case ImportOutcome.VALIDATED_ONLY: {
			return {
				message: ImportOutcomeMessage.VALIDATED_ONLY,
				tone: OutcomeTone.INFO,
			};
		}
	}
};

export { getOutcomeSummary };
