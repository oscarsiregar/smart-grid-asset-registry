import { ImportMode } from "~/modules/imports/imports.js";

import { ImportPanelMessage } from "../../../../libs/enums/enums.js";
import { type ModeOption } from "../types/types.js";

const MODE_OPTIONS: readonly ModeOption[] = [
	{
		description: ImportPanelMessage.MODE_PARTIAL_DESCRIPTION,
		label: ImportPanelMessage.MODE_PARTIAL_LABEL,
		value: ImportMode.PARTIAL,
	},
	{
		description: ImportPanelMessage.MODE_STRICT_DESCRIPTION,
		label: ImportPanelMessage.MODE_STRICT_LABEL,
		value: ImportMode.STRICT,
	},
];

export { MODE_OPTIONS };
