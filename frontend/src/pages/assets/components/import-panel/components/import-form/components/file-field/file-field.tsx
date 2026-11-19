import { useCallback, useId } from "react";
import {
	type Control,
	type FieldPath,
	type FieldValues,
	useController,
} from "react-hook-form";

import styles from "./styles.module.css";

type Properties<T extends FieldValues> = {
	accept: string;
	control: Control<T, null>;
	description: string;
	label: string;
	name: FieldPath<T>;
	onChange?: () => void;
};

/**
 * A labelled file picker bound to the form. A file input cannot be given a
 * value, so it reports the chosen File to the form instead of spreading the
 * field into the input.
 */
const FileField = <T extends FieldValues>({
	accept,
	control,
	description,
	label,
	name,
	onChange,
}: Properties<T>): React.JSX.Element => {
	const {
		field: {
			disabled: isDisabled,
			name: fieldName,
			onBlur: handleBlur,
			onChange: setValue,
			ref: inputReference,
		},
		fieldState: { error },
	} = useController({ control, name });

	const inputId = useId();
	const descriptionId = useId();
	const errorId = useId();
	const errorMessage = error?.message;

	const handleChange = useCallback(
		(event: React.ChangeEvent<HTMLInputElement>): void => {
			const [selectedFile] = event.target.files ?? [];
			setValue(selectedFile);
			onChange?.();
		},
		[onChange, setValue],
	);

	return (
		<div className={styles["field"]}>
			<label className={styles["label"]} htmlFor={inputId}>
				{label}
			</label>
			<input
				accept={accept}
				aria-describedby={
					errorMessage === undefined
						? descriptionId
						: `${descriptionId} ${errorId}`
				}
				aria-invalid={errorMessage !== undefined}
				className={styles["input"]}
				disabled={isDisabled}
				id={inputId}
				name={fieldName}
				onBlur={handleBlur}
				onChange={handleChange}
				ref={inputReference}
				type="file"
			/>
			<span className={styles["description"]} id={descriptionId}>
				{description}
			</span>
			{errorMessage !== undefined && (
				<span className={styles["error"]} id={errorId} role="alert">
					{errorMessage}
				</span>
			)}
		</div>
	);
};

export { FileField };
