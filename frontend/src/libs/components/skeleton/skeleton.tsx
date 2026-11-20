import { getValidClasses } from "~/libs/helpers/helpers.js";
import { type ValueOf } from "~/libs/types/types.js";

import { SkeletonShape, SkeletonWidth } from "./libs/enums/enums.js";
import styles from "./styles.module.css";

type Properties = {
	className?: string | undefined;
	shape?: ValueOf<typeof SkeletonShape>;
	width?: ValueOf<typeof SkeletonWidth>;
};

/**
 * One placeholder block standing in for content that is loading. It is
 * decorative: the area around it announces the loading text. It stays
 * transparent for a short delay, so a fast response never flashes it, then
 * fades in and pulses (no motion with reduced motion).
 */
const Skeleton: React.FC<Properties> = ({
	className,
	shape = SkeletonShape.TEXT,
	width = SkeletonWidth.FULL,
}: Properties) => (
	<span
		aria-hidden="true"
		className={getValidClasses(
			styles["skeleton"],
			styles[shape],
			styles[width],
			className,
		)}
		data-shape={shape}
		data-width={width}
	/>
);

export { Skeleton };
