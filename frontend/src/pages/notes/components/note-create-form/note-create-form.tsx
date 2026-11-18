import { useCallback } from "react";

import { Button } from "~/libs/components/button/button.js";
import { FormAlert } from "~/libs/components/form-alert/form-alert.js";
import { Input } from "~/libs/components/input/input.js";
import { NotificationType } from "~/libs/components/overlay-host/libs/enums/enums.js";
import { Textarea } from "~/libs/components/textarea/textarea.js";
import { IconName } from "~/libs/enums/enums.js";
import { useAppForm } from "~/libs/hooks/use-app-form/use-app-form.hook.js";
import { useServerFormErrors } from "~/libs/hooks/use-server-form-errors/use-server-form-errors.hook.js";
import { checkIsToastedError } from "~/libs/modules/api/libs/helpers/check-is-toasted-error.helper.js";
import { showNotification } from "~/libs/modules/notification/notification.js";
import { useCreateNoteMutation } from "~/modules/notes/notes-api.js";
import {
	type NoteCreateRequestDto,
	noteCreateValidationSchema,
} from "~/modules/notes/notes.js";

import { NotesPageMessage } from "../../libs/enums/enums.js";
import {
	DEFAULT_NOTE_CREATE_PAYLOAD,
	NOTE_CREATE_FIELDS,
} from "./libs/constants/constants.js";
import styles from "./styles.module.css";

const NoteCreateForm: React.FC = () => {
	const [createNote, { error, isLoading }] = useCreateNoteMutation();

	const { clearErrors, control, handleSubmit, reset, setError } =
		useAppForm<NoteCreateRequestDto>({
			defaultValues: DEFAULT_NOTE_CREATE_PAYLOAD,
			validationSchema: noteCreateValidationSchema,
		});

	const { hasFieldErrors } = useServerFormErrors({
		clearErrors,
		error,
		fields: NOTE_CREATE_FIELDS,
		setError,
	});
	const generalError =
		hasFieldErrors || checkIsToastedError(error) ? undefined : error;

	const handleCreate = useCallback(
		async (payload: NoteCreateRequestDto): Promise<void> => {
			try {
				await createNote(payload).unwrap();
				reset(DEFAULT_NOTE_CREATE_PAYLOAD);
				showNotification({
					message: NotesPageMessage.CREATE_SUCCESS,
					type: NotificationType.SUCCESS,
				});
			} catch {
				// The failure is exposed through the mutation error state and rendered above the form.
			}
		},
		[createNote, reset],
	);

	const handleFormSubmit = useCallback(
		(event: React.BaseSyntheticEvent): void => {
			void handleSubmit(handleCreate)(event);
		},
		[handleCreate, handleSubmit],
	);

	return (
		<section className={styles["card"]}>
			<h2 className={styles["title"]}>{NotesPageMessage.FORM_TITLE}</h2>
			<FormAlert error={generalError} />
			<form className={styles["form"]} noValidate onSubmit={handleFormSubmit}>
				<Input control={control} isRequired label="Title" name="title" />
				<Textarea control={control} label="Content" name="content" rows={4} />
				<Button
					className={styles["submit"]}
					iconName={IconName.PLUS}
					isLoading={isLoading}
					label={NotesPageMessage.SUBMIT_LABEL}
					type="submit"
				/>
			</form>
		</section>
	);
};

export { NoteCreateForm };
