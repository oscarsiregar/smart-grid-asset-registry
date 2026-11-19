import { fireEvent, render, screen } from "@testing-library/react";
import { useCallback, useRef, useState } from "react";
import { MemoryRouter, useLocation } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { Link } from "~/libs/components/link/link.js";

import { TabIndex } from "../../enums/enums.js";
import { getAssetPath } from "../../helpers/helpers.js";
import { useTreeKeyboard } from "./use-tree-keyboard.hook.js";

// A two-level tree: A (children A1, A2) and B (a leaf). The harness owns the
// expanded set like the real tree does; the hook only handles the keyboard.
const LocationProbe: React.FC = () => {
	const { pathname } = useLocation();

	return <output data-testid="pathname">{pathname}</output>;
};

type ItemProperties = {
	assetId: string;
	children?: React.ReactNode;
	focusedAssetId: null | string;
	isExpanded?: boolean | undefined;
	level: number;
};

const Item: React.FC<ItemProperties> = ({
	assetId,
	children,
	focusedAssetId,
	isExpanded,
	level,
}: ItemProperties) => (
	<li
		aria-expanded={isExpanded}
		aria-level={level}
		aria-selected={false}
		data-asset-id={assetId}
		role="treeitem"
		tabIndex={
			assetId === focusedAssetId ? TabIndex.FOCUSABLE : TabIndex.SKIPPED
		}
	>
		<Link tabIndex={-1} to={getAssetPath(assetId)}>
			{assetId}
		</Link>
		{children}
	</li>
);

type HarnessProperties = {
	preferredAssetId: null | string;
};

const Harness: React.FC<HarnessProperties> = ({
	preferredAssetId,
}: HarnessProperties) => {
	const treeReference = useRef<HTMLUListElement>(null);
	const [isAExpanded, setIsAExpanded] = useState(false);

	const handleExpand = useCallback((assetId: string): void => {
		if (assetId === "A") {
			setIsAExpanded(true);
		}
	}, []);

	const handleCollapse = useCallback((assetId: string): void => {
		if (assetId === "A") {
			setIsAExpanded(false);
		}
	}, []);

	const { focusedAssetId, handleFocus, handleKeyDown } = useTreeKeyboard({
		onCollapse: handleCollapse,
		onExpand: handleExpand,
		preferredAssetId,
		treeReference,
	});

	return (
		<ul
			aria-label="Assets"
			onFocus={handleFocus}
			onKeyDown={handleKeyDown}
			ref={treeReference}
			role="tree"
		>
			<Item
				assetId="A"
				focusedAssetId={focusedAssetId}
				isExpanded={isAExpanded}
				level={1}
			>
				{isAExpanded && (
					<ul role="group">
						<Item assetId="A1" focusedAssetId={focusedAssetId} level={2} />
						<Item assetId="A2" focusedAssetId={focusedAssetId} level={2} />
					</ul>
				)}
			</Item>
			<Item assetId="B" focusedAssetId={focusedAssetId} level={1} />
		</ul>
	);
};

const renderHarness = (preferredAssetId: null | string = null): void => {
	render(
		<MemoryRouter initialEntries={["/assets"]}>
			<Harness preferredAssetId={preferredAssetId} />
			<LocationProbe />
		</MemoryRouter>,
	);
};

const getItem = (assetId: string): HTMLElement => {
	const item = screen
		.getAllByRole("treeitem")
		.find((element) => element.dataset["assetId"] === assetId);

	if (!item) {
		throw new Error(`No tree item ${assetId}`);
	}

	return item;
};

const press = (key: string): void => {
	const focused = document.activeElement;

	if (!(focused instanceof HTMLElement)) {
		throw new TypeError("Nothing is focused");
	}

	fireEvent.keyDown(focused, { key });
};

const getTabStops = (): HTMLElement[] =>
	screen
		.getAllByRole("treeitem")
		.filter((element) => element.getAttribute("tabindex") === "0");

describe("useTreeKeyboard", () => {
	it("has exactly one tab stop, the first node by default", () => {
		renderHarness();

		expect(getTabStops()).toEqual([getItem("A")]);
	});

	it("puts the tab stop on the preferred (selected) node", () => {
		renderHarness("B");

		expect(getTabStops()).toEqual([getItem("B")]);
	});

	it("moves between visible nodes with the arrow keys, Home and End", () => {
		renderHarness();
		getItem("A").focus();

		press("ArrowDown");
		expect(getItem("B")).toHaveFocus();

		press("ArrowDown");
		expect(getItem("B")).toHaveFocus();

		press("ArrowUp");
		expect(getItem("A")).toHaveFocus();

		press("End");
		expect(getItem("B")).toHaveFocus();

		press("Home");
		expect(getItem("A")).toHaveFocus();
		expect(getTabStops()).toEqual([getItem("A")]);
	});

	it("expands with ArrowRight, then moves to the first child", () => {
		renderHarness();
		getItem("A").focus();

		press("ArrowRight");
		expect(getItem("A")).toHaveAttribute("aria-expanded", "true");
		expect(getItem("A")).toHaveFocus();

		press("ArrowRight");
		expect(getItem("A1")).toHaveFocus();

		press("ArrowDown");
		expect(getItem("A2")).toHaveFocus();
		expect(getTabStops()).toEqual([getItem("A2")]);
	});

	it("moves to the parent with ArrowLeft, then collapses", () => {
		renderHarness();
		getItem("A").focus();
		press("ArrowRight");
		press("ArrowRight");

		press("ArrowLeft");
		expect(getItem("A")).toHaveFocus();

		press("ArrowLeft");
		expect(getItem("A")).toHaveAttribute("aria-expanded", "false");
		expect(screen.queryByText("A1")).not.toBeInTheDocument();
	});

	it("keeps a tab stop when the focused node is collapsed away", () => {
		renderHarness();
		getItem("A").focus();
		press("ArrowRight");
		press("ArrowRight");
		expect(getTabStops()).toEqual([getItem("A1")]);

		fireEvent.keyDown(getItem("A"), { key: "ArrowLeft" });

		expect(getTabStops()).toEqual([getItem("A")]);
	});

	it("follows the node's link with Enter", () => {
		renderHarness();
		getItem("B").focus();

		press("Enter");

		expect(screen.getByTestId("pathname")).toHaveTextContent("/assets/B");
	});
});
