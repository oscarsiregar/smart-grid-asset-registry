import { type z } from "zod";

import { type noteCreate } from "../validation-schemas/note-create.validation-schema.js";

type NoteCreateRequestDto = z.infer<typeof noteCreate>;

export { type NoteCreateRequestDto };
