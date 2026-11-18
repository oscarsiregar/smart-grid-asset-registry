import { NoteValidationRule } from "./note-validation-rule.enum.js";

const NoteValidationMessage = {
	CONTENT_TOO_LONG: `Content must be at most ${String(NoteValidationRule.CONTENT_MAXIMUM_LENGTH)} characters`,
	TITLE_REQUIRED: "Title is required",
	TITLE_TOO_LONG: `Title must be at most ${String(NoteValidationRule.TITLE_MAXIMUM_LENGTH)} characters`,
} as const;

export { NoteValidationMessage };
