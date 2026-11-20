import { SkeletonWidth } from "~/libs/components/skeleton/libs/enums/enums.js";

/** Widths that placeholder rows take in turn, so a list of them looks like text. */
const SKELETON_ROW_WIDTHS = [
	SkeletonWidth.THREE_QUARTERS,
	SkeletonWidth.HALF,
	SkeletonWidth.FULL,
	SkeletonWidth.THREE_QUARTERS,
] as const;

export { SKELETON_ROW_WIDTHS };
