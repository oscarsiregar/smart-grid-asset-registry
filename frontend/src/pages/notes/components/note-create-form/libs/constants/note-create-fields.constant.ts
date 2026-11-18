import { type FieldPath } from "react-hook-form";

import { type NoteCreateRequestDto } from "~/modules/notes/notes.js";

const NOTE_CREATE_FIELDS: FieldPath<NoteCreateRequestDto>[] = [
	"content",
	"title",
];

export { NOTE_CREATE_FIELDS };
