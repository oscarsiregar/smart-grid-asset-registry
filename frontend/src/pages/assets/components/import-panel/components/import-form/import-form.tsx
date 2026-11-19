import { useCallback } from "react";
import { useWatch } from "react-hook-form";

import { Button } from "~/libs/components/button/button.js";
import { useAppForm } from "~/libs/hooks/use-app-form/use-app-form.hook.js";
import {
	importFileValidationSchema,
	type ImportRequestDto,
} from "~/modules/imports/imports.js";

import { ImportPanelMessage } from "../../libs/enums/enums.js";
import { FileField } from "./components/file-field/file-field.js";
import { ModeField } from "./components/mode-field/mode-field.js";
import {
	ACCEPTED_FILE_TYPES,
	DEFAULT_IMPORT_PAYLOAD,
	MODE_OPTIONS,
} from "./libs/constants/constants.js";
import styles from "./styles.module.css";

type Properties = {
	isLoading: boolean;
	onChange: () => void;
	onSubmit: (payload: ImportRequestDto) => void;
};

/**
 * Chooses a file and a mode. The schema only checks that a file is chosen;
 * every rule about the file is the backend's (constitution, principle II).
 * Changing the file or the mode calls onChange so a stale result is cleared.
 */
const ImportForm: React.FC<Properties> = ({
	isLoading,
	onChange,
	onSubmit,
}: Properties) => {
	const { control, handleSubmit } = useAppForm<ImportRequestDto>({
		defaultValues: DEFAULT_IMPORT_PAYLOAD,
		validationSchema: importFileValidationSchema,
	});
	const file = useWatch({ control, name: "file" });

	const handleFormSubmit = useCallback(
		(event: React.BaseSyntheticEvent): void => {
			void handleSubmit(onSubmit)(event);
		},
		[handleSubmit, onSubmit],
	);

	return (
		<form className={styles["form"]} noValidate onSubmit={handleFormSubmit}>
			<FileField
				accept={ACCEPTED_FILE_TYPES}
				control={control}
				description={ImportPanelMessage.FILE_DESCRIPTION}
				label={ImportPanelMessage.FILE_LABEL}
				name="file"
				onChange={onChange}
			/>
			<ModeField
				control={control}
				legend={ImportPanelMessage.MODE_LEGEND}
				name="mode"
				onChange={onChange}
				options={MODE_OPTIONS}
			/>
			<div className={styles["actions"]}>
				<Button
					isDisabled={!file}
					isLoading={isLoading}
					label={ImportPanelMessage.CHECK_BUTTON_LABEL}
					type="submit"
				/>
				{isLoading && (
					<span className={styles["status"]} role="status">
						{ImportPanelMessage.CHECKING_STATUS}
					</span>
				)}
			</div>
		</form>
	);
};

export { ImportForm };
