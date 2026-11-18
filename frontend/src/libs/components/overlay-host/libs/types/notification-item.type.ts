import { type ValueOf } from "~/libs/types/types.js";

import { type NotificationType } from "../enums/enums.js";

type NotificationItem = {
	id: string;
	isClosing: boolean;
	message: string;
	type: ValueOf<typeof NotificationType>;
};

export { type NotificationItem };
