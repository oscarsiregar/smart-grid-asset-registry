import { Icon } from "~/libs/components/icon/icon.js";
import { Link } from "~/libs/components/link/link.js";
import { EMPTY_LENGTH } from "~/libs/constants/constants.js";
import { IconName } from "~/libs/enums/enums.js";
import { configureString } from "~/libs/helpers/helpers.js";
import { type AssetNodeDto } from "~/modules/assets/assets.js";

import { AssetsPageMessage } from "../../../../libs/enums/enums.js";
import { getAssetPath } from "../../../../libs/helpers/helpers.js";
import { IdTag } from "../../../id-tag/id-tag.js";
import { StatusBadge } from "../../../status-badge/status-badge.js";
import styles from "./styles.module.css";

type Properties = {
	childAssets: AssetNodeDto[];
	label: string;
};

/**
 * Children under a heading with their number, each a row with name, ID tag
 * and status badge, in the order received; an empty list says "None"
 * (spec FR-012, FR-014, FR-015). The name link is stretched over the row, so
 * the whole row opens the asset while there is still one link per row.
 */
const ChildList: React.FC<Properties> = ({
	childAssets,
	label,
}: Properties) => (
	<section className={styles["section"]}>
		<h4 className={styles["heading"]}>
			{configureString(AssetsPageMessage.GROUP_HEADING, {
				count: String(childAssets.length),
				label,
			})}
		</h4>
		{childAssets.length === EMPTY_LENGTH ? (
			<p className={styles["empty"]}>{AssetsPageMessage.GROUP_EMPTY}</p>
		) : (
			<ul className={styles["list"]}>
				{childAssets.map((child) => (
					<li className={styles["item"]} key={child.assetId}>
						<Link
							className={styles["link"]}
							hasDefaultStyles={false}
							to={getAssetPath(child.assetId)}
						>
							<span className={styles["name"]}>{child.assetName}</span>
							<IdTag assetId={child.assetId} />
						</Link>
						<StatusBadge status={child.operationalStatus} />
						<Icon className={styles["chevron"]} iconName={IconName.CHEVRON} />
					</li>
				))}
			</ul>
		)}
	</section>
);

export { ChildList };
