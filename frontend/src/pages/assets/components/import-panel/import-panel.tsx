import { useCallback, useState } from "react";

import { FormAlert } from "~/libs/components/form-alert/form-alert.js";
import {
	useCreateImportMutation,
	useValidateImportMutation,
} from "~/modules/imports/imports-api.js";
import { type ImportRequestDto } from "~/modules/imports/imports.js";

import { ImportForm } from "./components/import-form/import-form.js";
import { ImportResult } from "./components/import-result/import-result.js";
import { OutcomeBanner } from "./components/outcome-banner/outcome-banner.js";
import { FORM_KEY_STEP, INITIAL_FORM_KEY } from "./libs/constants/constants.js";
import {
	ImportOutcomeMessage,
	ImportPanelMessage,
	OutcomeTone,
} from "./libs/enums/enums.js";
import { checkIsFailure } from "./libs/helpers/helpers.js";
import styles from "./styles.module.css";

/**
 * Check a CSV file (user story 1), then save it (user story 2). Results come
 * from the mutation states, never a copy; only the file and mode that were
 * checked are kept, so Save sends exactly that file again and the server
 * re-validates it (spec FR-022).
 */
const ImportPanel: React.FC = () => {
	const [
		validateImport,
		{
			data: validated,
			error: checkError,
			isLoading: isChecking,
			reset: resetValidate,
		},
	] = useValidateImportMutation();
	const [
		createImport,
		{
			data: confirmed,
			error: saveError,
			isLoading: isSaving,
			reset: resetCreate,
		},
	] = useCreateImportMutation();
	const [checkedPayload, setCheckedPayload] = useState<ImportRequestDto | null>(
		null,
	);
	const [formKey, setFormKey] = useState(INITIAL_FORM_KEY);

	const handleSubmit = useCallback(
		(payload: ImportRequestDto): void => {
			setCheckedPayload(payload);
			void validateImport(payload);
		},
		[validateImport],
	);

	const handleConfirm = useCallback((): void => {
		if (checkedPayload) {
			void createImport(checkedPayload);
		}
	}, [checkedPayload, createImport]);

	const handleChange = useCallback((): void => {
		resetValidate();
	}, [resetValidate]);

	const handleReset = useCallback((): void => {
		resetValidate();
		resetCreate();
		setCheckedPayload(null);
		setFormKey((key) => key + FORM_KEY_STEP);
	}, [resetCreate, resetValidate]);

	if (confirmed) {
		return (
			<ImportResult
				onReset={handleReset}
				resetLabel={ImportPanelMessage.IMPORT_ANOTHER_FILE_LABEL}
				result={confirmed}
			/>
		);
	}

	if (validated) {
		const hasSaveFailed = checkIsFailure(saveError);

		return (
			<div className={styles["panel"]}>
				{hasSaveFailed && (
					<OutcomeBanner
						message={ImportOutcomeMessage.COMMIT_FAILED}
						tone={OutcomeTone.DANGER}
					/>
				)}
				<FormAlert error={hasSaveFailed ? undefined : saveError} />
				<ImportResult
					isSaving={isSaving}
					onConfirm={handleConfirm}
					onReset={handleReset}
					resetLabel={ImportPanelMessage.CHOOSE_ANOTHER_FILE_LABEL}
					result={validated}
				/>
			</div>
		);
	}

	const hasCheckFailed = checkIsFailure(checkError);

	return (
		<div className={styles["panel"]}>
			<ImportForm
				isLoading={isChecking}
				key={formKey}
				onChange={handleChange}
				onSubmit={handleSubmit}
			/>
			<FormAlert error={hasCheckFailed ? undefined : checkError} />
			{hasCheckFailed && (
				<OutcomeBanner
					message={ImportOutcomeMessage.CHECK_FAILED}
					tone={OutcomeTone.DANGER}
				/>
			)}
		</div>
	);
};

export { ImportPanel };
