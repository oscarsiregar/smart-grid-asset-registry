import { type ImportIssueDto } from "./import-issue-dto.type.js";

type ImportRejectedRowDto = {
	assetId: null | string;
	errors: ImportIssueDto[];
	row: number;
};

export { type ImportRejectedRowDto };
