import { type ValueOf } from "~/libs/types/types.js";
import { OperationalStatus } from "~/modules/assets/assets.js";

import styles from "./styles.module.css";

type Properties = {
	status: null | ValueOf<typeof OperationalStatus>;
};

/**
 * The shape that goes with a status, so the status is never shown by colour
 * alone (constitution principle V): filled for in service, half-filled for
 * maintenance, an open ring for out of service and a dashed ring for no
 * status. It takes its colour from the badge (currentColor).
 */
const StatusShape: React.FC<Properties> = ({ status }: Properties) => (
	<svg
		aria-hidden="true"
		className={styles["shape"]}
		focusable="false"
		viewBox="0 0 12 12"
	>
		{status === OperationalStatus.IN_SERVICE && (
			<circle cx="6" cy="6" fill="currentColor" r="5" />
		)}
		{status === OperationalStatus.MAINTENANCE && (
			<>
				<circle
					cx="6"
					cy="6"
					fill="none"
					r="4.25"
					stroke="currentColor"
					strokeWidth="1.5"
				/>
				<path d="M6 1.75a4.25 4.25 0 0 1 0 8.5z" fill="currentColor" />
			</>
		)}
		{status === OperationalStatus.OUT_OF_SERVICE && (
			<circle
				cx="6"
				cy="6"
				fill="none"
				r="4.25"
				stroke="currentColor"
				strokeWidth="1.5"
			/>
		)}
		{status === null && (
			<circle
				cx="6"
				cy="6"
				fill="none"
				r="4.25"
				stroke="currentColor"
				strokeDasharray="2 1.6"
				strokeWidth="1.5"
			/>
		)}
	</svg>
);

export { StatusShape };
