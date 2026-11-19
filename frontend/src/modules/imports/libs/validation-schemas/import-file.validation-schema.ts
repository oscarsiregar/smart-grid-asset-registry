import { z } from "zod";

import { ImportMode, ImportValidationMessage } from "../enums/enums.js";

const importFile = z
	.object({
		file: z.instanceof(File, { error: ImportValidationMessage.FILE_REQUIRED }),
		mode: z.enum(ImportMode, { error: ImportValidationMessage.MODE_REQUIRED }),
	})
	.required();

export { importFile };
