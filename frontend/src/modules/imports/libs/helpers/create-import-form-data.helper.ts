import { ImportFormField } from "../enums/enums.js";
import { type ImportRequestDto } from "../types/types.js";

const createImportFormData = ({ file, mode }: ImportRequestDto): FormData => {
	const formData = new FormData();
	formData.append(ImportFormField.FILE, file);
	formData.append(ImportFormField.MODE, mode);

	return formData;
};

export { createImportFormData };
