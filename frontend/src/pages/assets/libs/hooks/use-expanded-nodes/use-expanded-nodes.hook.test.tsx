import { act, renderHook } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { useExpandedNodes } from "./use-expanded-nodes.hook.js";

type Properties = {
	selectedAssetId: null | string;
	selectedPath: null | readonly string[];
};

const PANEL_PATH = ["SS-001", "SWB-001-1", "PNL-001-1-01"] as const;

const renderExpandedNodes = (initialProperties: Properties) =>
	renderHook((properties: Properties) => useExpandedNodes(properties), {
		initialProps: initialProperties,
	});

const getExpanded = (
	result: ReturnType<typeof renderExpandedNodes>["result"],
): string[] =>
	[...result.current.expandedAssetIds].toSorted((first, second) =>
		first.localeCompare(second),
	);

describe("useExpandedNodes", () => {
	it("expands nothing until the selected asset's path arrives", () => {
		const { result } = renderExpandedNodes({
			selectedAssetId: "PNL-001-1-01",
			selectedPath: null,
		});

		expect(getExpanded(result)).toEqual([]);
	});

	it("expands the selected asset's ancestors, not the asset itself, in the same render", () => {
		const { rerender, result } = renderExpandedNodes({
			selectedAssetId: "PNL-001-1-01",
			selectedPath: null,
		});

		rerender({ selectedAssetId: "PNL-001-1-01", selectedPath: PANEL_PATH });

		expect(getExpanded(result)).toEqual(["SS-001", "SWB-001-1"]);
	});

	it("ignores a path that belongs to a previous selection", () => {
		const { result } = renderExpandedNodes({
			selectedAssetId: "SS-002",
			selectedPath: PANEL_PATH,
		});

		expect(getExpanded(result)).toEqual([]);
	});

	it("lets the user collapse an ancestor afterwards without re-expanding it", () => {
		const { rerender, result } = renderExpandedNodes({
			selectedAssetId: "PNL-001-1-01",
			selectedPath: PANEL_PATH,
		});

		act(() => {
			result.current.collapse("SWB-001-1");
		});
		rerender({ selectedAssetId: "PNL-001-1-01", selectedPath: PANEL_PATH });

		expect(getExpanded(result)).toEqual(["SS-001"]);
	});

	it("adds the ancestors of a new selection and keeps what was expanded", () => {
		const { rerender, result } = renderExpandedNodes({
			selectedAssetId: "PNL-001-1-01",
			selectedPath: PANEL_PATH,
		});

		rerender({ selectedAssetId: "TX-002-1", selectedPath: PANEL_PATH });
		rerender({
			selectedAssetId: "TX-002-1",
			selectedPath: ["SS-002", "TX-002-1"],
		});

		expect(getExpanded(result)).toEqual(["SS-001", "SS-002", "SWB-001-1"]);
	});

	it("expands, collapses and toggles single nodes", () => {
		const { result } = renderExpandedNodes({
			selectedAssetId: null,
			selectedPath: null,
		});

		act(() => {
			result.current.expand("SS-001");
		});
		act(() => {
			result.current.toggle("SS-002");
		});
		expect(getExpanded(result)).toEqual(["SS-001", "SS-002"]);

		act(() => {
			result.current.toggle("SS-001");
		});
		expect(getExpanded(result)).toEqual(["SS-002"]);
	});
});
