import { Button } from "~/libs/components/button/button.js";
import { Link } from "~/libs/components/link/link.js";
import { AppRoute, ButtonVariant, ControlSize } from "~/libs/enums/enums.js";

import { AssetsPageMessage } from "../../../../libs/enums/enums.js";
import styles from "./styles.module.css";

type Properties = {
	/** Offer the way back to the overview (the not-found state, spec FR-020). */
	hasBackLink?: boolean;
	message: string;
	onRetry?: (() => void) | undefined;
};

/**
 * A details-area state in words: the prompt, not found with a way back, or a
 * failure with a retry. Loading shows placeholders instead (spec 005).
 */
const DetailsMessage: React.FC<Properties> = ({
	hasBackLink = false,
	message,
	onRetry,
}: Properties) => (
	<div className={styles["message"]}>
		<p className={styles["text"]}>{message}</p>
		{hasBackLink && (
			<Link to={AppRoute.ASSETS}>{AssetsPageMessage.BACK_TO_ALL_ASSETS}</Link>
		)}
		{onRetry && (
			<Button
				label={AssetsPageMessage.TRY_AGAIN}
				onClick={onRetry}
				size={ControlSize.SM}
				type="button"
				variant={ButtonVariant.SECONDARY}
			/>
		)}
	</div>
);

export { DetailsMessage };
