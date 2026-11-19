import { EMPTY_LENGTH } from "~/libs/constants/constants.js";
import { configureString, getValidClasses } from "~/libs/helpers/helpers.js";
import { type ImportResultDto } from "~/modules/imports/imports.js";

import { ImportPanelMessage } from "../../libs/enums/enums.js";
import styles from "./styles.module.css";

type Properties = {
	warnings: ImportResultDto["warnings"];
};

/** Warnings never block a row (spec FR-020); they are listed apart from errors. */
const WarningsTable: React.FC<Properties> = ({ warnings }: Properties) => {
	if (warnings.length === EMPTY_LENGTH) {
		return <p className={styles["empty"]}>{ImportPanelMessage.NO_WARNINGS}</p>;
	}

	return (
		<table className={styles["table"]}>
			<caption className={styles["caption"]}>
				{configureString(ImportPanelMessage.WARNINGS_CAPTION, {
					count: String(warnings.length),
				})}
			</caption>
			<thead>
				<tr>
					<th
						className={getValidClasses(styles["cell"], styles["number"])}
						scope="col"
					>
						{ImportPanelMessage.TABLE_ROW}
					</th>
					<th className={styles["cell"]} scope="col">
						{ImportPanelMessage.TABLE_ASSET_ID}
					</th>
					<th className={styles["cell"]} scope="col">
						{ImportPanelMessage.TABLE_CODE}
					</th>
					<th className={styles["cell"]} scope="col">
						{ImportPanelMessage.TABLE_MESSAGE}
					</th>
				</tr>
			</thead>
			<tbody>
				{warnings.map((warning) => (
					<tr key={`${String(warning.row)}-${warning.code}`}>
						<td className={getValidClasses(styles["cell"], styles["number"])}>
							{warning.row}
						</td>
						<td className={styles["cell"]}>{warning.assetId}</td>
						<td className={styles["cell"]}>
							<code className={styles["code"]}>{warning.code}</code>
						</td>
						<td className={styles["cell"]}>{warning.message}</td>
					</tr>
				))}
			</tbody>
		</table>
	);
};

export { WarningsTable };
