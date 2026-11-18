import { z } from "zod";

import { NoteValidationMessage, NoteValidationRule } from "../enums/enums.js";

const noteCreate = z
	.object({
		content: z.string().trim().max(NoteValidationRule.CONTENT_MAXIMUM_LENGTH, {
			error: NoteValidationMessage.CONTENT_TOO_LONG,
		}),
		title: z
			.string()
			.trim()
			.min(NoteValidationRule.TITLE_EMPTY_STATE_LENGTH, {
				error: NoteValidationMessage.TITLE_REQUIRED,
			})
			.max(NoteValidationRule.TITLE_MAXIMUM_LENGTH, {
				error: NoteValidationMessage.TITLE_TOO_LONG,
			}),
	})
	.required();

export { noteCreate };
