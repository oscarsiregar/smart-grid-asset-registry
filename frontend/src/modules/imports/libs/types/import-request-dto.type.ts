import { type z } from "zod";

import { type importFile } from "../validation-schemas/import-file.validation-schema.js";

type ImportRequestDto = z.infer<typeof importFile>;

export { type ImportRequestDto };
