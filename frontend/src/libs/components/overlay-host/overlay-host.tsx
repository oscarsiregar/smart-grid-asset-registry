import { useCallback, useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";

import { Notification } from "~/libs/components/notification/notification.js";
import { EMPTY_LENGTH } from "~/libs/constants/constants.js";
import { EventType, KeyboardKey } from "~/libs/enums/enums.js";
import {
	bindShowNotification,
	type ShowNotificationPayload,
} from "~/libs/modules/notification/notification.js";

import {
	DEFAULT_NOTIFICATION_DURATION_MS,
	DEFAULT_NOTIFICATION_TYPE,
	NOTIFICATION_CLOSING_DURATION_MS,
} from "./libs/constants/constants.js";
import { generateNotificationId } from "./libs/helpers/helpers.js";
import { type NotificationItem } from "./libs/types/types.js";
import styles from "./styles.module.css";

type Properties = {
	children: React.ReactNode;
};

const OverlayHost = ({ children }: Properties) => {
	const [notifications, setNotifications] = useState<NotificationItem[]>([]);
	const notificationIdCounterReference = useRef(EMPTY_LENGTH);
	const timeoutIdsReference = useRef<
		Map<string, ReturnType<typeof setTimeout>>
	>(new Map());

	const dismissNotification = useCallback((id: string) => {
		const timeoutId = timeoutIdsReference.current.get(id);

		if (timeoutId !== undefined) {
			clearTimeout(timeoutId);
			timeoutIdsReference.current.delete(id);
		}

		setNotifications((currentNotifications) => {
			return currentNotifications.filter((item) => item.id !== id);
		});
	}, []);

	const scheduleNotificationDismiss = useCallback(
		(notification: NotificationItem) => {
			if (notification.isClosing) {
				return;
			}

			const previousTimeoutId = timeoutIdsReference.current.get(
				notification.id,
			);

			if (previousTimeoutId !== undefined) {
				clearTimeout(previousTimeoutId);
			}

			const timeoutId = setTimeout(() => {
				dismissNotification(notification.id);
			}, NOTIFICATION_CLOSING_DURATION_MS);

			timeoutIdsReference.current.set(notification.id, timeoutId);
		},
		[dismissNotification],
	);

	const dismissNotifications = useCallback(() => {
		for (const notification of notifications) {
			scheduleNotificationDismiss(notification);
		}

		setNotifications((currentNotifications) =>
			currentNotifications.map((notification) => ({
				...notification,
				isClosing: true,
			})),
		);
	}, [scheduleNotificationDismiss, notifications]);

	const handleShowNotification = useCallback(
		(payload: ShowNotificationPayload) => {
			const id =
				payload.id ?? generateNotificationId(notificationIdCounterReference);

			if (timeoutIdsReference.current.has(id)) {
				return;
			}

			const duration = payload.duration ?? DEFAULT_NOTIFICATION_DURATION_MS;
			const type = payload.type ?? DEFAULT_NOTIFICATION_TYPE;

			setNotifications((currentNotifications) => [
				...currentNotifications,
				{
					id,
					isClosing: false,
					message: payload.message,
					type,
				},
			]);

			const timeoutId = setTimeout(() => {
				dismissNotification(id);
			}, duration);

			timeoutIdsReference.current.set(id, timeoutId);
		},
		[dismissNotification],
	);

	useEffect(() => {
		const handleKeyPress = (event: KeyboardEvent) => {
			if (event.key === KeyboardKey.ESCAPE) {
				dismissNotifications();
			}
		};

		document.addEventListener(EventType.KEYDOWN, handleKeyPress);

		return () => {
			document.removeEventListener(EventType.KEYDOWN, handleKeyPress);
		};
	}, [dismissNotifications]);

	useEffect(() => {
		return bindShowNotification(handleShowNotification);
	}, [handleShowNotification]);

	useEffect(() => {
		const timeoutIds = timeoutIdsReference.current;

		return () => {
			for (const timeoutId of timeoutIds.values()) {
				clearTimeout(timeoutId);
			}

			timeoutIds.clear();
		};
	}, []);

	return (
		<>
			{children}
			{createPortal(
				<div
					aria-atomic="false"
					aria-live="polite"
					className={styles["notifications"]}
				>
					{notifications.map((item) => (
						<Notification item={item} key={item.id} />
					))}
				</div>,
				document.body,
			)}
		</>
	);
};

export { OverlayHost };
