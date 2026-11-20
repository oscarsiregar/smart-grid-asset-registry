import { Button } from "~/libs/components/button/button.js";
import { ButtonVariant } from "~/libs/enums/enums.js";
import { configureString } from "~/libs/helpers/helpers.js";

import { ImportPanelMessage } from "../../libs/enums/enums.js";
import styles from "./styles.module.css";

type Properties = {
	/** The number of new assets a save would store. */
	confirmCount: number;
	isSaving?: boolean;
	onConfirm?: (() => void) | undefined;
	onReset: () => void;
	resetLabel: string;
};

/**
 * What can be done with a result: Save (a checked result only) and start
 * again, with "Saving…" while a save runs. Shown in the drawer's footer
 * (spec 005 FR-018).
 */
const ImportResultActions: React.FC<Properties> = ({
	confirmCount,
	isSaving = false,
	onConfirm,
	onReset,
	resetLabel,
}: Properties) => (
	<div className={styles["actions"]}>
		{onConfirm && (
			<Button
				isDisabled={isSaving}
				isLoading={isSaving}
				label={configureString(ImportPanelMessage.CONFIRM_BUTTON_LABEL, {
					count: String(confirmCount),
				})}
				onClick={onConfirm}
				type="button"
			/>
		)}
		<Button
			isDisabled={isSaving}
			label={resetLabel}
			onClick={onReset}
			type="button"
			variant={ButtonVariant.SECONDARY}
		/>
		{isSaving && (
			<span className={styles["status"]} role="status">
				{ImportPanelMessage.SAVING_STATUS}
			</span>
		)}
	</div>
);

export { ImportResultActions };
