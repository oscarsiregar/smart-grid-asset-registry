import { type AssetNodeDto } from "~/modules/assets/assets.js";

import { CHILD_GROUPS } from "../../../../libs/constants/constants.js";
import { ChildList } from "../child-list/child-list.js";
import styles from "./styles.module.css";

type Properties = {
	childAssets: AssetNodeDto[];
};

/**
 * A substation's children in three groups, always all three (spec FR-012).
 * Filtering keeps the order the API sent; nothing is re-sorted here
 * (principle II).
 */
const ChildGroups: React.FC<Properties> = ({ childAssets }: Properties) => (
	<div className={styles["groups"]}>
		{CHILD_GROUPS.map(({ assetType, label }) => (
			<ChildList
				childAssets={childAssets.filter(
					(child) => child.assetType === assetType,
				)}
				key={assetType}
				label={label}
			/>
		))}
	</div>
);

export { ChildGroups };
