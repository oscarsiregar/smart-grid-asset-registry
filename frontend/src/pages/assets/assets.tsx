import { Button } from "~/libs/components/button/button.js";
import { PageContainer } from "~/libs/components/page-container/page-container.js";
import { PageIntro } from "~/libs/components/page-intro/page-intro.js";
import { Panel } from "~/libs/components/panel/panel.js";
import { IconName, PanelName } from "~/libs/enums/enums.js";
import { usePanel } from "~/libs/hooks/use-panel/use-panel.hook.js";

import { AssetsPageMessage } from "./libs/enums/enums.js";
import styles from "./styles.module.css";

const Assets: React.FC = () => {
	const { closePanel, isOpen, openPanel } = usePanel(PanelName.IMPORT);

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
				<p className={styles["empty"]}>{AssetsPageMessage.EMPTY}</p>
			</div>
			<Panel
				isOpen={isOpen}
				onClose={closePanel}
				title={AssetsPageMessage.IMPORT_PANEL_TITLE}
			>
				<p className={styles["empty"]}>
					{AssetsPageMessage.IMPORT_PANEL_PLACEHOLDER}
				</p>
			</Panel>
		</PageContainer>
	);
};

export { Assets };
