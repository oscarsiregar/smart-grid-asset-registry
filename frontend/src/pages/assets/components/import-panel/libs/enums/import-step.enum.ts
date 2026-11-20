/** The steps of an import, as the step indicator shows them (spec 005 FR-017). */
const ImportStep = {
	CHOOSE_FILE: "choose-file",
	DONE: "done",
	REVIEW: "review",
} as const;

export { ImportStep };
