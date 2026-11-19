import { useParams } from "react-router-dom";

import { Button } from "~/libs/components/button/button.js";
import { PageContainer } from "~/libs/components/page-container/page-container.js";
import { PageIntro } from "~/libs/components/page-intro/page-intro.js";
import { Panel } from "~/libs/components/panel/panel.js";
import { EMPTY_LENGTH } from "~/libs/constants/constants.js";
import { IconName, PanelName } from "~/libs/enums/enums.js";
import { usePanel } from "~/libs/hooks/use-panel/use-panel.hook.js";
import { useGetAssetNodesQuery } from "~/modules/assets/assets.js";

import { AssetExplorer } from "./components/asset-explorer/asset-explorer.js";
import { EmptyRegistry } from "./components/empty-registry/empty-registry.js";
import { ImportPanel } from "./components/import-panel/import-panel.js";
import { AssetsPageMessage } from "./libs/enums/enums.js";
import styles from "./styles.module.css";

/**
 * The asset explorer. The selected asset is the `:assetId` path segment
 * (decoded by the router); the import panel is `?panel=import`.
 */
const Assets: React.FC = () => {
	const { assetId } = useParams();
	const { closePanel, isOpen, openPanel } = usePanel(PanelName.IMPORT);
	const { data } = useGetAssetNodesQuery(null);

	const selectedAssetId =
		assetId === undefined || assetId === "" ? null : assetId;
	const isRegistryEmpty = data?.totalItems === EMPTY_LENGTH;

	return (
		<PageContainer>
			<div className={styles["content"]}>
				<PageIntro
					description={AssetsPageMessage.DESCRIPTION}
					label={AssetsPageMessage.KICKER}
					title={AssetsPageMessage.TITLE}
				/>
				<div className={styles["actions"]}>
					<Button
						iconName={IconName.PLUS}
						label={AssetsPageMessage.IMPORT_BUTTON_LABEL}
						onClick={openPanel}
						type="button"
					/>
				</div>
				{isRegistryEmpty ? (
					<EmptyRegistry />
				) : (
					<AssetExplorer selectedAssetId={selectedAssetId} />
				)}
			</div>
			<Panel
				isOpen={isOpen}
				onClose={closePanel}
				title={AssetsPageMessage.IMPORT_PANEL_TITLE}
			>
				<ImportPanel />
			</Panel>
		</PageContainer>
	);
};

export { Assets };
