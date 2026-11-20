import { useCallback } from "react";

import { Button } from "~/libs/components/button/button.js";
import { Input } from "~/libs/components/input/input.js";
import { ButtonVariant, ControlSize, IconName } from "~/libs/enums/enums.js";
import { useAppForm } from "~/libs/hooks/use-app-form/use-app-form.hook.js";
import {
	type AssetSearchRequestDto,
	assetSearchValidationSchema,
} from "~/modules/assets/assets.js";

import { AssetsPageMessage } from "../../../../libs/enums/enums.js";
import { AssetTypeField } from "./components/asset-type-field/asset-type-field.js";
import { EMPTY_SEARCH } from "./libs/constants/constants.js";
import styles from "./styles.module.css";

type Properties = {
	/** The search in the address; the form starts with it. */
	activeSearch: AssetSearchRequestDto | null;
	onClear: () => void;
	onSubmit: (search: AssetSearchRequestDto) => void;
};

/**
 * The search form. The schema only refuses a blank text (spec FR-004); the
 * length limit and matching are the backend's (principle II). The parent
 * remounts the form when the address's search changes, so its values follow
 * Back, Forward and refresh without an effect.
 */
const SearchForm: React.FC<Properties> = ({
	activeSearch,
	onClear,
	onSubmit,
}: Properties) => {
	const { control, handleSubmit } = useAppForm<AssetSearchRequestDto>({
		defaultValues: activeSearch ?? EMPTY_SEARCH,
		validationSchema: assetSearchValidationSchema,
	});

	// Only the values go up; React Hook Form would also pass the event.
	const handleValidSubmit = useCallback(
		(values: AssetSearchRequestDto): void => {
			onSubmit(values);
		},
		[onSubmit],
	);

	const handleFormSubmit = useCallback(
		(event: React.BaseSyntheticEvent): void => {
			void handleSubmit(handleValidSubmit)(event);
		},
		[handleSubmit, handleValidSubmit],
	);

	return (
		<form
			aria-label={AssetsPageMessage.SEARCH_FORM_LABEL}
			className={styles["form"]}
			noValidate
			onSubmit={handleFormSubmit}
			role="search"
		>
			<div className={styles["field"]}>
				<Input
					control={control}
					iconName={IconName.SEARCH}
					isLabelHidden
					label={AssetsPageMessage.SEARCH_FIELD_LABEL}
					name="q"
					placeholder={AssetsPageMessage.SEARCH_FIELD_PLACEHOLDER}
				/>
			</div>
			<AssetTypeField control={control} name="assetType" />
			<div className={styles["actions"]}>
				<Button
					label={AssetsPageMessage.SEARCH_BUTTON_LABEL}
					size={ControlSize.SM}
					type="submit"
					variant={ButtonVariant.SECONDARY}
				/>
				{activeSearch && (
					<Button
						label={AssetsPageMessage.CLEAR_SEARCH_BUTTON_LABEL}
						onClick={onClear}
						size={ControlSize.SM}
						type="button"
						variant={ButtonVariant.SECONDARY}
					/>
				)}
			</div>
		</form>
	);
};

export { SearchForm };
