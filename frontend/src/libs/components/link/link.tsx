import { useCallback } from "react";
import { NavLink } from "react-router-dom";

import { NEW_TAB_LINK_ATTRIBUTES } from "~/libs/constants/constants.js";
import { getValidClasses } from "~/libs/helpers/helpers.js";
import {
	type GeneratedRoutePath,
	type NavigableRoute,
} from "~/libs/types/types.js";

import styles from "./styles.module.css";

type Properties = {
	activeClassName?: string | undefined;
	children: React.ReactNode;
	className?: string | undefined;
	hasDefaultStyles?: boolean;
	shouldOpenInNewTab?: boolean;
	/** For composite widgets (a tree) that manage focus themselves. */
	tabIndex?: number | undefined;
	/** A fixed route, or an address with parameters from `generateRoutePath`. */
	to: GeneratedRoutePath | NavigableRoute;
};

const Link: React.FC<Properties> = ({
	activeClassName,
	children,
	className,
	hasDefaultStyles = true,
	shouldOpenInNewTab = false,
	tabIndex,
	to,
}: Properties) => {
	const getLinkClassName = useCallback(
		({ isActive }: { isActive: boolean }): string =>
			getValidClasses(
				hasDefaultStyles && styles["link"],
				hasDefaultStyles && isActive && styles["active"],
				className,
				isActive && activeClassName,
			),
		[activeClassName, className, hasDefaultStyles],
	);

	const newTabAttributes = shouldOpenInNewTab ? NEW_TAB_LINK_ATTRIBUTES : {};

	return (
		<NavLink
			{...newTabAttributes}
			className={getLinkClassName}
			tabIndex={tabIndex}
			to={to}
		>
			{children}
		</NavLink>
	);
};

export { Link };
