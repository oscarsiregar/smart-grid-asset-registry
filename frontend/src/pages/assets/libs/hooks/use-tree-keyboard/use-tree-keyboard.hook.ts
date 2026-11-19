import { useCallback, useEffect, useState } from "react";

import { KeyboardKey } from "~/libs/enums/enums.js";

import { TREE_ITEM_SELECTOR, TREE_STEP } from "../../constants/constants.js";

type Parameters = {
	onCollapse: (assetId: string) => void;
	onExpand: (assetId: string) => void;
	/** The node Tab should land on until another node takes focus (the selection). */
	preferredAssetId: null | string;
	treeReference: React.RefObject<HTMLElement | null>;
};

type TreeKeyboard = {
	/** The single node with `tabIndex={0}` (roving tab index). */
	focusedAssetId: null | string;
	handleFocus: (event: React.FocusEvent<HTMLElement>) => void;
	handleKeyDown: (event: React.KeyboardEvent<HTMLElement>) => void;
};

const getItems = (tree: HTMLElement): HTMLElement[] => [
	...tree.querySelectorAll<HTMLElement>(TREE_ITEM_SELECTOR),
];

const getAssetId = (item: HTMLElement): null | string =>
	item.dataset["assetId"] ?? null;

const getParentItem = (item: HTMLElement): HTMLElement | null =>
	item.parentElement?.closest<HTMLElement>(TREE_ITEM_SELECTOR) ?? null;

const getItemOf = (target: EventTarget): HTMLElement | null =>
	target instanceof Element
		? target.closest<HTMLElement>(TREE_ITEM_SELECTOR)
		: null;

/**
 * Keyboard support of the WAI-ARIA navigation tree (spec FR-007, research
 * R8): the nodes are read from the DOM in document order, which is the
 * visible order because collapsed children are not rendered.
 */
const useTreeKeyboard = ({
	onCollapse,
	onExpand,
	preferredAssetId,
	treeReference,
}: Parameters): TreeKeyboard => {
	const [focusedState, setFocusedState] = useState<null | string>(null);
	const [previousPreferredAssetId, setPreviousPreferredAssetId] =
		useState(preferredAssetId);

	// A new selection becomes the tab stop again (adjusting state during
	// render, not in an effect).
	if (previousPreferredAssetId !== preferredAssetId) {
		setPreviousPreferredAssetId(preferredAssetId);
		setFocusedState(null);
	}

	const focusedAssetId = focusedState ?? preferredAssetId;

	// The tab stop may not be rendered (collapsed away, or not loaded yet):
	// then the first node takes it, so the tree stays reachable with Tab.
	useEffect(() => {
		const tree = treeReference.current;

		if (!tree) {
			return;
		}

		const items = getItems(tree);
		const [firstItem] = items;
		const hasTabStop = items.some(
			(item) => getAssetId(item) === focusedAssetId,
		);

		if (!hasTabStop && firstItem) {
			setFocusedState(getAssetId(firstItem));
		}
	});

	const handleFocus = useCallback(
		(event: React.FocusEvent<HTMLElement>): void => {
			const item = getItemOf(event.target);

			if (item) {
				setFocusedState(getAssetId(item));
			}
		},
		[],
	);

	const handleKeyDown = useCallback(
		(event: React.KeyboardEvent<HTMLElement>): void => {
			const tree = treeReference.current;
			const current = getItemOf(event.target);
			const assetId = current ? getAssetId(current) : null;

			if (!tree || !current || assetId === null) {
				return;
			}

			const items = getItems(tree);
			const position = items.indexOf(current);
			const expandedState = current.getAttribute("aria-expanded");
			let target: HTMLElement | null | undefined = null;

			switch (event.key) {
				case KeyboardKey.ARROW_DOWN: {
					target = items[position + TREE_STEP];
					break;
				}

				case KeyboardKey.ARROW_LEFT: {
					if (expandedState === "true") {
						onCollapse(assetId);
					} else {
						target = getParentItem(current);
					}

					break;
				}

				case KeyboardKey.ARROW_RIGHT: {
					const next = items[position + TREE_STEP];

					if (expandedState === "false") {
						onExpand(assetId);
					} else if (next && getParentItem(next) === current) {
						target = next;
					}

					break;
				}

				case KeyboardKey.ARROW_UP: {
					target = items[position - TREE_STEP];
					break;
				}

				case KeyboardKey.END: {
					target = items[items.length - TREE_STEP];
					break;
				}

				case KeyboardKey.ENTER: {
					// Only on the node itself: a focused link or button handles
					// Enter natively.
					if (event.target !== current) {
						return;
					}

					current.querySelector("a")?.click();
					break;
				}

				case KeyboardKey.HOME: {
					[target] = items;
					break;
				}

				default: {
					return;
				}
			}

			event.preventDefault();
			target?.focus();
		},
		[onCollapse, onExpand, treeReference],
	);

	return { focusedAssetId, handleFocus, handleKeyDown };
};

export { useTreeKeyboard };
