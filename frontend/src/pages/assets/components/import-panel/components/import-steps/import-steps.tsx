import { Icon } from "~/libs/components/icon/icon.js";
import { IconName } from "~/libs/enums/enums.js";
import { getValidClasses } from "~/libs/helpers/helpers.js";
import { type ValueOf } from "~/libs/types/types.js";

import {
	IMPORT_STEPS,
	STEP_NUMBER_OFFSET,
} from "../../libs/constants/constants.js";
import { ImportPanelMessage, type ImportStep } from "../../libs/enums/enums.js";
import styles from "./styles.module.css";

type Properties = {
	currentStep: ValueOf<typeof ImportStep>;
};

/**
 * Where the import is: Choose file, Review, Done (spec 005 FR-017). A display
 * of the flow, not a navigation. The current step is outlined and announced
 * as current; earlier steps show a check and "(done)" for screen readers.
 */
const ImportSteps: React.FC<Properties> = ({ currentStep }: Properties) => {
	const currentPosition = IMPORT_STEPS.findIndex(
		({ step }) => step === currentStep,
	);

	return (
		<ol aria-label={ImportPanelMessage.STEPS_LABEL} className={styles["steps"]}>
			{IMPORT_STEPS.map(({ label, step }, position) => {
				const isCurrent = position === currentPosition;
				const isDone = position < currentPosition;

				return (
					<li
						aria-current={isCurrent ? "step" : undefined}
						className={getValidClasses(
							styles["step"],
							isCurrent && styles["current"],
							isDone && styles["done"],
						)}
						key={step}
					>
						<span className={styles["number"]}>
							{position + STEP_NUMBER_OFFSET}
						</span>
						{label}
						{isDone && (
							<>
								<Icon className={styles["check"]} iconName={IconName.CHECK} />
								<span className="visually-hidden">
									{ImportPanelMessage.STEP_DONE_MARKER}
								</span>
							</>
						)}
					</li>
				);
			})}
		</ol>
	);
};

export { ImportSteps };
