import { useCallback, useEffect, useId, useState } from "react";
import { useLocation, useParams } from "react-router-dom";

import { Button } from "~/libs/components/button/button.js";
import { Icon } from "~/libs/components/icon/icon.js";
import { Logo } from "~/libs/components/logo/logo.js";
import { Panel } from "~/libs/components/panel/panel.js";
import { EMPTY_LENGTH } from "~/libs/constants/constants.js";
import {
	ButtonVariant,
	ControlSize,
	IconName,
	KeyboardKey,
	PanelName,
} from "~/libs/enums/enums.js";
import { getValidClasses } from "~/libs/helpers/helpers.js";
import { usePanel } from "~/libs/hooks/use-panel/use-panel.hook.js";
import { useGetAssetNodesQuery } from "~/modules/assets/assets.js";

import { AssetDetails } from "./components/asset-details/asset-details.js";
import { AssetExplorer } from "./components/asset-explorer/asset-explorer.js";
import { EmptyRegistry } from "./components/empty-registry/empty-registry.js";
import { ImportPanel } from "./components/import-panel/import-panel.js";
import { AssetsPageMessage, TabIndex } from "./libs/enums/enums.js";
import styles from "./styles.module.css";

/**
 * The asset explorer. The selected asset is the `:assetId` path segment
 * (decoded by the router); the import panel is `?panel=import`.
 *
 * Layout (spec 004 FR-001, FR-002): a sidebar with the logo, the import
 * action, the search and the tree, beside a sticky header and the details.
 * Below 960 px the sidebar is an off-canvas drawer; it is open only for the
 * location it was opened on, so any navigation (selecting an asset, a search)
 * closes it without an effect.
 */
const Assets: React.FC = () => {
	const { assetId } = useParams();
	const location = useLocation();
	const { closePanel, isOpen, openPanel } = usePanel(PanelName.IMPORT);
	const { data } = useGetAssetNodesQuery(null);
	const [sidebarOpenedAt, setSidebarOpenedAt] = useState<null | string>(null);
	const sidebarId = useId();

	const selectedAssetId =
		assetId === undefined || assetId === "" ? null : assetId;
	const isRegistryEmpty = data?.totalItems === EMPTY_LENGTH;
	const isSidebarOpen = sidebarOpenedAt === location.key;

	const handleSidebarOpen = useCallback((): void => {
		setSidebarOpenedAt(location.key);
	}, [location.key]);

	const handleSidebarClose = useCallback((): void => {
		setSidebarOpenedAt(null);
	}, []);

	useEffect(() => {
		if (!isSidebarOpen) {
			return;
		}

		const handleKeyDown = (event: KeyboardEvent): void => {
			if (event.key === KeyboardKey.ESCAPE) {
				setSidebarOpenedAt(null);
			}
		};

		document.addEventListener("keydown", handleKeyDown);

		return (): void => {
			document.removeEventListener("keydown", handleKeyDown);
		};
	}, [isSidebarOpen]);

	return (
		<div className={styles["layout"]}>
			{isSidebarOpen && (
				<button
					aria-label={AssetsPageMessage.CLOSE_MENU_LABEL}
					className={styles["scrim"]}
					onClick={handleSidebarClose}
					tabIndex={TabIndex.SKIPPED}
					type="button"
				/>
			)}
			<aside
				aria-label={AssetsPageMessage.NAVIGATION_LABEL}
				className={getValidClasses(
					styles["sidebar"],
					isSidebarOpen && styles["sidebar-open"],
				)}
				id={sidebarId}
			>
				<div className={styles["brand"]}>
					<Logo size={ControlSize.SM} />
				</div>
				<Button
					className={styles["import-button"]}
					iconName={IconName.PLUS}
					label={AssetsPageMessage.IMPORT_BUTTON_LABEL}
					onClick={openPanel}
					type="button"
					variant={ButtonVariant.SECONDARY}
				/>
				{!isRegistryEmpty && (
					<AssetExplorer selectedAssetId={selectedAssetId} />
				)}
			</aside>
			<div className={styles["body"]}>
				<header className={styles["header"]}>
					<button
						aria-controls={sidebarId}
						aria-expanded={isSidebarOpen}
						aria-label={AssetsPageMessage.OPEN_MENU_LABEL}
						className={styles["menu-button"]}
						onClick={handleSidebarOpen}
						type="button"
					>
						<Icon className={styles["menu-icon"]} iconName={IconName.MENU} />
					</button>
					<div className={styles["header-copy"]}>
						<h1 className={styles["title"]}>{AssetsPageMessage.TITLE}</h1>
						<p className={styles["subtitle"]}>
							{AssetsPageMessage.DESCRIPTION}
						</p>
					</div>
				</header>
				<div className={styles["content"]}>
					{isRegistryEmpty ? (
						<EmptyRegistry />
					) : (
						<AssetDetails selectedAssetId={selectedAssetId} />
					)}
				</div>
			</div>
			<Panel
				isOpen={isOpen}
				onClose={closePanel}
				title={AssetsPageMessage.IMPORT_PANEL_TITLE}
			>
				<ImportPanel />
			</Panel>
		</div>
	);
};

export { Assets };
