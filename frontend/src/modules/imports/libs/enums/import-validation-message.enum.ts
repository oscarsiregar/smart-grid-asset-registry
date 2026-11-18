/**
 * Messages for the import form. The form only checks that a file is chosen and
 * the mode is known; every rule about the file itself is the backend's
 * (constitution, principle II).
 */
const ImportValidationMessage = {
	FILE_REQUIRED: "Choose a CSV file.",
	MODE_REQUIRED: "Choose an import mode.",
} as const;

export { ImportValidationMessage };
