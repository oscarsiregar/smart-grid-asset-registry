import { SkeletonWidth } from "~/libs/components/skeleton/libs/enums/enums.js";
import { type ValueOf } from "~/libs/types/types.js";

import { SKELETON_ROW_WIDTHS } from "../constants/constants.js";

/** The width of the placeholder row at a position, taken in turn. */
const getSkeletonRowWidth = (position: number): ValueOf<typeof SkeletonWidth> =>
	SKELETON_ROW_WIDTHS[position % SKELETON_ROW_WIDTHS.length] ??
	SkeletonWidth.FULL;

export { getSkeletonRowWidth };
