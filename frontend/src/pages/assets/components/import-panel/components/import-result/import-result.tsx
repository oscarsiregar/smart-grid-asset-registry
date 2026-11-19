import { Button } from "~/libs/components/button/button.js";
import { ButtonVariant } from "~/libs/enums/enums.js";
import { configureString } from "~/libs/helpers/helpers.js";
import { type ImportResultDto } from "~/modules/imports/imports.js";

import { ImportPanelMessage } from "../../libs/enums/enums.js";
import { getOutcomeSummary } from "../../libs/helpers/helpers.js";
import { OutcomeBanner } from "../outcome-banner/outcome-banner.js";
import { RejectedRowsTable } from "../rejected-rows-table/rejected-rows-table.js";
import { WarningsTable } from "../warnings-table/warnings-table.js";
import styles from "./styles.module.css";

type Properties = {
	isSaving?: boolean;
	onConfirm?: (() => void) | undefined;
	onReset: () => void;
	resetLabel: string;
	result: ImportResultDto;
};

/**
 * One import result: the outcome in words, the counts, the rejected rows and
 * the warnings. A validated result offers Save (onConfirm); a confirmed one
 * only offers to start again.
 */
const ImportResult: React.FC<Properties> = ({
	isSaving = false,
	onConfirm,
	onReset,
	resetLabel,
	result,
}: Properties) => {
	const { counts, isCommitted, rejectedRows, warnings } = result;
	const importedLabel = isCommitted
		? ImportPanelMessage.COUNT_IMPORTED
		: ImportPanelMessage.COUNT_TO_IMPORT;

	return (
		<div className={styles["result"]}>
			<OutcomeBanner {...getOutcomeSummary(result)} />
			<dl
				aria-label={ImportPanelMessage.COUNTS_LABEL}
				className={styles["counts"]}
			>
				<div className={styles["count"]}>
					<dt className={styles["count-label"]}>
						{ImportPanelMessage.COUNT_TOTAL}
					</dt>
					<dd className={styles["count-value"]}>{counts.total}</dd>
				</div>
				<div className={styles["count"]}>
					<dt className={styles["count-label"]}>{importedLabel}</dt>
					<dd className={styles["count-value"]}>{counts.imported}</dd>
				</div>
				<div className={styles["count"]}>
					<dt className={styles["count-label"]}>
						{ImportPanelMessage.COUNT_UNCHANGED}
					</dt>
					<dd className={styles["count-value"]}>{counts.unchanged}</dd>
				</div>
				<div className={styles["count"]}>
					<dt className={styles["count-label"]}>
						{ImportPanelMessage.COUNT_REJECTED}
					</dt>
					<dd className={styles["count-value"]}>{counts.rejected}</dd>
				</div>
				<div className={styles["count"]}>
					<dt className={styles["count-label"]}>
						{ImportPanelMessage.COUNT_WARNING_ROWS}
					</dt>
					<dd className={styles["count-value"]}>{counts.warningRows}</dd>
				</div>
			</dl>
			<RejectedRowsTable rows={rejectedRows} />
			<WarningsTable warnings={warnings} />
			<div className={styles["actions"]}>
				{onConfirm && (
					<Button
						isDisabled={isSaving}
						isLoading={isSaving}
						label={configureString(ImportPanelMessage.CONFIRM_BUTTON_LABEL, {
							count: String(counts.imported),
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
		</div>
	);
};

export { ImportResult };
