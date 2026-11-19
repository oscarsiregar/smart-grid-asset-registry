/** Outcome of validating or confirming a file (contracts/imports-api.md). */
const ImportOutcome = {
	COMMITTED: "COMMITTED",
	NOTHING_TO_COMMIT: "NOTHING_TO_COMMIT",
	STRICT_BLOCKED: "STRICT_BLOCKED",
	VALIDATED_ONLY: "VALIDATED_ONLY",
} as const;

export { ImportOutcome };
