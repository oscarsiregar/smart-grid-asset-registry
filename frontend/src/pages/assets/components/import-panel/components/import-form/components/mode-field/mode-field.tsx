import { useCallback, useId } from "react";
import {
	type Control,
	type FieldPath,
	type FieldValues,
	useController,
} from "react-hook-form";

import { type ModeOption } from "../../libs/types/types.js";
import styles from "./styles.module.css";

type Properties<T extends FieldValues> = {
	control: Control<T, null>;
	legend: string;
	name: FieldPath<T>;
	onChange?: () => void;
	options: readonly ModeOption[];
};

/** A radio group with a one-line explanation per option. */
const ModeField = <T extends FieldValues>({
	control,
	legend,
	name,
	onChange,
	options,
}: Properties<T>): React.JSX.Element => {
	const { field } = useController({ control, name });
	const groupId = useId();

	const handleChange = useCallback(
		(event: React.ChangeEvent<HTMLInputElement>): void => {
			field.onChange(event.target.value);
			onChange?.();
		},
		[field, onChange],
	);

	return (
		<fieldset className={styles["fieldset"]} disabled={field.disabled}>
			<legend className={styles["legend"]}>{legend}</legend>
			{options.map((option) => {
				const descriptionId = `${groupId}-${option.value}`;

				return (
					<label className={styles["option"]} key={option.value}>
						<input
							aria-describedby={descriptionId}
							checked={field.value === option.value}
							className={styles["radio"]}
							name={field.name}
							onBlur={field.onBlur}
							onChange={handleChange}
							type="radio"
							value={option.value}
						/>
						<span className={styles["option-label"]}>{option.label}</span>
						<span className={styles["option-description"]} id={descriptionId}>
							{option.description}
						</span>
					</label>
				);
			})}
		</fieldset>
	);
};

export { ModeField };
