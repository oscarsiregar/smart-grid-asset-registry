/**
 * The commit outcome every result states in words (spec FR-027, constitution
 * principle III). `:count` is filled in by getOutcomeSummary.
 */
const ImportOutcomeMessage = {
	CHECK_FAILED:
		"Not checked — the file could not be checked. Nothing was saved.",
	COMMIT_FAILED: "Not committed — the import failed and was rolled back",
	COMMITTED: "Committed — :count assets saved",
	NOTHING_TO_COMMIT: "Nothing to commit — no new assets",
	STRICT_BLOCKED: "Not committed — strict mode, :count rows rejected",
	VALIDATED_ONLY: "Validated only — nothing saved",
} as const;

export { ImportOutcomeMessage };
