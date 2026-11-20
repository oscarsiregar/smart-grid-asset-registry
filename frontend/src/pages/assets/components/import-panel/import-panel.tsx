import { useCallback, useState } from "react";

import { FormAlert } from "~/libs/components/form-alert/form-alert.js";
import {
	useCreateImportMutation,
	useValidateImportMutation,
} from "~/modules/imports/imports-api.js";
import { type ImportRequestDto } from "~/modules/imports/imports.js";

import { ImportForm } from "./components/import-form/import-form.js";
import { ImportResultActions } from "./components/import-result-actions/import-result-actions.js";
import { ImportResultSkeleton } from "./components/import-result-skeleton/import-result-skeleton.js";
import { ImportResult } from "./components/import-result/import-result.js";
import { ImportSteps } from "./components/import-steps/import-steps.js";
import { OutcomeBanner } from "./components/outcome-banner/outcome-banner.js";
import { FORM_KEY_STEP, INITIAL_FORM_KEY } from "./libs/constants/constants.js";
import {
	ImportOutcomeMessage,
	ImportPanelMessage,
	OutcomeTone,
} from "./libs/enums/enums.js";
import { checkIsFailure, getImportStep } from "./libs/helpers/helpers.js";
import styles from "./styles.module.css";

/**
 * Check a CSV file (user story 1), then save it (user story 2). Results come
 * from the mutation states, never a copy; only the file and mode that were
 * checked are kept, so Save sends exactly that file again and the server
 * re-validates it (spec FR-022). The steps sit at the top; the content
 * scrolls, and the actions stay in a footer below it (spec 005 FR-017,
 * FR-018).
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

	const hasCheckFailed = checkIsFailure(checkError);
	const hasSaveFailed = checkIsFailure(saveError);

	const renderContent = (): React.ReactNode => {
		if (confirmed) {
			return <ImportResult result={confirmed} />;
		}

		if (validated) {
			return (
				<>
					{hasSaveFailed && (
						<OutcomeBanner
							message={ImportOutcomeMessage.COMMIT_FAILED}
							tone={OutcomeTone.DANGER}
						/>
					)}
					<FormAlert error={hasSaveFailed ? undefined : saveError} />
					<ImportResult result={validated} />
				</>
			);
		}

		return (
			<>
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
				{isChecking && <ImportResultSkeleton />}
			</>
		);
	};

	const renderActions = (): React.ReactNode => {
		if (confirmed) {
			return (
				<ImportResultActions
					confirmCount={confirmed.counts.imported}
					onReset={handleReset}
					resetLabel={ImportPanelMessage.IMPORT_ANOTHER_FILE_LABEL}
				/>
			);
		}

		if (validated) {
			return (
				<ImportResultActions
					confirmCount={validated.counts.imported}
					isSaving={isSaving}
					onConfirm={handleConfirm}
					onReset={handleReset}
					resetLabel={ImportPanelMessage.CHOOSE_ANOTHER_FILE_LABEL}
				/>
			);
		}

		return null;
	};

	const actions = renderActions();

	return (
		<div className={styles["panel"]}>
			<div className={styles["content"]} data-import-content>
				<ImportSteps currentStep={getImportStep({ confirmed, validated })} />
				{renderContent()}
			</div>
			{actions && (
				<footer className={styles["footer"]} data-import-footer>
					{actions}
				</footer>
			)}
		</div>
	);
};

export { ImportPanel };
