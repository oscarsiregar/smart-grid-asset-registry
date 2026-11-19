import { useCallback, useId } from "react";
import {
	type Control,
	type FieldPath,
	type FieldValues,
	useController,
} from "react-hook-form";

import { ASSET_TYPE_FILTER_OPTIONS } from "../../../../../../libs/constants/constants.js";
import {
	AssetsPageMessage,
	AssetTypeFilter,
} from "../../../../../../libs/enums/enums.js";
import { assetTypeToLabel } from "../../../../../../libs/maps.js";
import styles from "./styles.module.css";

type Properties<T extends FieldValues> = {
	control: Control<T, null>;
	name: FieldPath<T>;
};

/**
 * The type a search is limited to. "All types" is `null` in the form's
 * values (the schema's all types); the select shows it as `AssetTypeFilter.ALL`.
 */
const AssetTypeField = <T extends FieldValues>({
	control,
	name,
}: Properties<T>): React.JSX.Element => {
	const { field } = useController({ control, name });
	const selectId = useId();

	const handleChange = useCallback(
		(event: React.ChangeEvent<HTMLSelectElement>): void => {
			const { value } = event.target;
			field.onChange(value === AssetTypeFilter.ALL ? null : value);
		},
		[field],
	);

	return (
		<div className={styles["field"]}>
			<label className={styles["label"]} htmlFor={selectId}>
				{AssetsPageMessage.ASSET_TYPE_FIELD_LABEL}
			</label>
			<select
				className={styles["select"]}
				disabled={field.disabled}
				id={selectId}
				name={field.name}
				onBlur={field.onBlur}
				onChange={handleChange}
				value={field.value ?? AssetTypeFilter.ALL}
			>
				<option value={AssetTypeFilter.ALL}>
					{AssetsPageMessage.ALL_TYPES}
				</option>
				{ASSET_TYPE_FILTER_OPTIONS.map((assetType) => (
					<option key={assetType} value={assetType}>
						{assetTypeToLabel[assetType]}
					</option>
				))}
			</select>
		</div>
	);
};

export { AssetTypeField };
