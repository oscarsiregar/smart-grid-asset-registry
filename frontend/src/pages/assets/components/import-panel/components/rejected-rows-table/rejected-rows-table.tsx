import { EMPTY_LENGTH } from "~/libs/constants/constants.js";
import { configureString, getValidClasses } from "~/libs/helpers/helpers.js";
import { type ImportResultDto } from "~/modules/imports/imports.js";

import { ImportPanelMessage } from "../../libs/enums/enums.js";
import styles from "./styles.module.css";

type Properties = {
	rows: ImportResultDto["rejectedRows"];
};

/** One table line per error, so a row with several errors shows them all. */
const RejectedRowsTable: React.FC<Properties> = ({ rows }: Properties) => {
	if (rows.length === EMPTY_LENGTH) {
		return (
			<p className={styles["empty"]}>{ImportPanelMessage.NO_REJECTED_ROWS}</p>
		);
	}

	return (
		<table className={styles["table"]}>
			<caption className={styles["caption"]}>
				{configureString(ImportPanelMessage.REJECTED_CAPTION, {
					count: String(rows.length),
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
				{rows.flatMap((row) =>
					row.errors.map((error) => (
						<tr key={`${String(row.row)}-${error.code}`}>
							<td className={getValidClasses(styles["cell"], styles["number"])}>
								{row.row}
							</td>
							<td className={styles["cell"]}>
								{row.assetId ?? (
									<>
										<span aria-hidden="true">
											{ImportPanelMessage.EMPTY_CELL}
										</span>
										<span className="visually-hidden">
											{ImportPanelMessage.ASSET_ID_NONE}
										</span>
									</>
								)}
							</td>
							<td className={styles["cell"]}>
								<code className={styles["code"]}>{error.code}</code>
							</td>
							<td className={styles["cell"]}>{error.message}</td>
						</tr>
					)),
				)}
			</tbody>
		</table>
	);
};

export { RejectedRowsTable };
