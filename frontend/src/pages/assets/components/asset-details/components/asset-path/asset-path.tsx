import { Link } from "~/libs/components/link/link.js";
import { type AssetNodeDto } from "~/modules/assets/assets.js";

import { AssetsPageMessage } from "../../../../libs/enums/enums.js";
import { getAssetPath } from "../../../../libs/helpers/helpers.js";
import styles from "./styles.module.css";

type Properties = {
	currentAssetId: string;
	path: AssetNodeDto[];
};

/** The chain from the substation down to the asset; every ancestor is a link (spec FR-011). */
const AssetPath: React.FC<Properties> = ({
	currentAssetId,
	path,
}: Properties) => (
	<nav aria-label={AssetsPageMessage.PATH_LABEL}>
		<ol className={styles["path"]}>
			{path.map((step) => (
				<li className={styles["step"]} key={step.assetId}>
					{step.assetId === currentAssetId ? (
						<span aria-current="page" className={styles["current"]}>
							{step.assetId}
						</span>
					) : (
						<Link to={getAssetPath(step.assetId)}>{step.assetId}</Link>
					)}
				</li>
			))}
		</ol>
	</nav>
);

export { AssetPath };
