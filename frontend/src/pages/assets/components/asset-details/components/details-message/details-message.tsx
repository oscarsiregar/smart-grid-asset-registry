import { Button } from "~/libs/components/button/button.js";
import { ButtonVariant, ControlSize } from "~/libs/enums/enums.js";

import { AssetsPageMessage } from "../../../../libs/enums/enums.js";
import styles from "./styles.module.css";

type Properties = {
	isLoading?: boolean;
	message: string;
	onRetry?: (() => void) | undefined;
};

/** A details-area state in words: the prompt, loading, or a failure with a retry. */
const DetailsMessage: React.FC<Properties> = ({
	isLoading = false,
	message,
	onRetry,
}: Properties) => (
	<div className={styles["message"]}>
		{isLoading ? (
			<p className={styles["text"]} role="status">
				{message}
			</p>
		) : (
			<p className={styles["text"]}>{message}</p>
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
