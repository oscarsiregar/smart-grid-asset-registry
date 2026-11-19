import { Link } from "~/libs/components/link/link.js";
import { EMPTY_LENGTH } from "~/libs/constants/constants.js";
import { configureString } from "~/libs/helpers/helpers.js";
import { type AssetNodeDto } from "~/modules/assets/assets.js";

import { AssetsPageMessage } from "../../../../libs/enums/enums.js";
import {
	getAssetPath,
	getStatusText,
} from "../../../../libs/helpers/helpers.js";
import styles from "./styles.module.css";

type Properties = {
	childAssets: AssetNodeDto[];
	label: string;
};

/**
 * Children under a heading with their number, each a link with ID, name and
 * status in words, in the order received; an empty list says "None"
 * (spec FR-012, FR-014, FR-015).
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
						<Link className={styles["link"]} to={getAssetPath(child.assetId)}>
							<span className={styles["id"]}>{child.assetId}</span>
							<span>{child.assetName}</span>
						</Link>
						<span className={styles["status"]}>
							{getStatusText(child.operationalStatus)}
						</span>
					</li>
				))}
			</ul>
		)}
	</section>
);

export { ChildList };
