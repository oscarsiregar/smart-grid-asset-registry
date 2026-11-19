import { type DescendantCountsDto } from "~/modules/assets/assets.js";

import { AssetsPageMessage } from "../../../../libs/enums/enums.js";
import styles from "./styles.module.css";

type Properties = {
	counts: DescendantCountsDto;
};

/** How many of each type sit beneath a substation, on every level (spec FR-013). */
const DescendantCounts: React.FC<Properties> = ({ counts }: Properties) => {
	const rows = [
		{ label: AssetsPageMessage.COUNT_TRANSFORMERS, value: counts.transformers },
		{ label: AssetsPageMessage.COUNT_LV_BOARDS, value: counts.lvBoards },
		{ label: AssetsPageMessage.COUNT_SWITCHBOARDS, value: counts.switchboards },
		{
			label: AssetsPageMessage.COUNT_SWITCHBOARD_PANELS,
			value: counts.switchboardPanels,
		},
	];

	return (
		<section className={styles["section"]}>
			<h4 className={styles["heading"]}>{AssetsPageMessage.COUNTS_LABEL}</h4>
			<dl className={styles["counts"]}>
				{rows.map(({ label, value }) => (
					<div className={styles["count"]} key={label}>
						<dt className={styles["label"]}>{label}</dt>
						<dd className={styles["value"]}>{value}</dd>
					</div>
				))}
			</dl>
		</section>
	);
};

export { DescendantCounts };
