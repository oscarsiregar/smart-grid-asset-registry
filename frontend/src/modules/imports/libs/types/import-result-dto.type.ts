import { type ValueOf } from "~/libs/types/types.js";

import { type ImportMode, type ImportOutcome } from "../enums/enums.js";
import { type ImportCountsDto } from "./import-counts-dto.type.js";
import { type ImportRecordDto } from "./import-record-dto.type.js";
import { type ImportRejectedRowDto } from "./import-rejected-row-dto.type.js";
import { type ImportWarningDto } from "./import-warning-dto.type.js";

/** `ImportResult` from contracts/imports-api.md. */
type ImportResultDto = {
	counts: ImportCountsDto;
	fileName: string;
	import: ImportRecordDto | null;
	isCommitted: boolean;
	mode: ValueOf<typeof ImportMode>;
	outcome: ValueOf<typeof ImportOutcome>;
	rejectedRows: ImportRejectedRowDto[];
	warnings: ImportWarningDto[];
};

export { type ImportResultDto };
