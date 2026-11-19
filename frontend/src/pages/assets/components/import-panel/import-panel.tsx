import { useCallback, useState } from "react";

import { FormAlert } from "~/libs/components/form-alert/form-alert.js";
import { ErrorCode } from "~/libs/enums/enums.js";
import { checkIsServerError } from "~/libs/modules/api/libs/helpers/check-is-server-error.helper.js";
import { useValidateImportMutation } from "~/modules/imports/imports-api.js";
import { type ImportRequestDto } from "~/modules/imports/imports.js";

import { ImportForm } from "./components/import-form/import-form.js";
import { ImportResult } from "./components/import-result/import-result.js";
import { OutcomeBanner } from "./components/outcome-banner/outcome-banner.js";
import { FORM_KEY_STEP, INITIAL_FORM_KEY } from "./libs/constants/constants.js";
import { ImportOutcomeMessage, OutcomeTone } from "./libs/enums/enums.js";
import styles from "./styles.module.css";

/**
 * Checks a CSV file against the store (user story 1). The result comes from
 * the mutation state, never a copy. A file-level problem (4xx with an error
 * code) is shown as an alert with the server's message; a server or network
 * failure says that nothing was saved.
 */
const ImportPanel: React.FC = () => {
	const [validateImport, { data: result, error, isLoading, reset }] =
		useValidateImportMutation();
	const [formKey, setFormKey] = useState(INITIAL_FORM_KEY);

	const hasFailed =
		error !== undefined &&
		(!checkIsServerError(error) ||
			error.code === ErrorCode.INTERNAL_SERVER_ERROR);
	const fileError = hasFailed ? undefined : error;

	const handleSubmit = useCallback(
		(payload: ImportRequestDto): void => {
			void validateImport(payload);
		},
		[validateImport],
	);

	const handleChange = useCallback((): void => {
		reset();
	}, [reset]);

	const handleReset = useCallback((): void => {
		reset();
		setFormKey((key) => key + FORM_KEY_STEP);
	}, [reset]);

	if (result) {
		return <ImportResult onReset={handleReset} result={result} />;
	}

	return (
		<div className={styles["panel"]}>
			<ImportForm
				isLoading={isLoading}
				key={formKey}
				onChange={handleChange}
				onSubmit={handleSubmit}
			/>
			<FormAlert error={fileError} />
			{hasFailed && (
				<OutcomeBanner
					message={ImportOutcomeMessage.CHECK_FAILED}
					tone={OutcomeTone.DANGER}
				/>
			)}
		</div>
	);
};

export { ImportPanel };
