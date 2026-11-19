import { IconName } from "~/libs/enums/enums.js";
import { type ValueOf } from "~/libs/types/types.js";

import { OutcomeTone } from "../../../libs/enums/enums.js";

const outcomeToneToIconName = {
	[OutcomeTone.DANGER]: IconName.CIRCLE_X,
	[OutcomeTone.INFO]: IconName.INFO,
	[OutcomeTone.SUCCESS]: IconName.CHECK_CIRCLE,
	[OutcomeTone.WARNING]: IconName.ALERT_CIRCLE,
} satisfies Record<ValueOf<typeof OutcomeTone>, ValueOf<typeof IconName>>;

export { outcomeToneToIconName };
