import { act, renderHook } from "@testing-library/react";
import { useCallback } from "react";
import { MemoryRouter, useLocation, useNavigate } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { AssetType } from "~/modules/assets/assets.js";

import { useAssetSearch } from "./use-asset-search.hook.js";

const HISTORY_BACK = -1;

// The hook plus what the address looks like and a way back.
const useHarness = () => {
	const navigate = useNavigate();
	const { search } = useLocation();
	const goBack = useCallback((): void => {
		void navigate(HISTORY_BACK);
	}, [navigate]);

	return { goBack, search, ...useAssetSearch() };
};

const renderAt = (initialEntry: string) =>
	renderHook(() => useHarness(), {
		wrapper: ({ children }: { children: React.ReactNode }) => (
			<MemoryRouter initialEntries={[initialEntry]}>{children}</MemoryRouter>
		),
	});

describe("useAssetSearch", () => {
	it("has no search without q or with a blank q", () => {
		expect(renderAt("/assets").result.current.activeSearch).toBeNull();
		expect(renderAt("/assets?q=%20%20").result.current.activeSearch).toBeNull();
	});

	it("reads the text and the type from the address", () => {
		const { result } = renderAt("/assets?q=PNL-001&type=SWITCHBOARD_PANEL");

		expect(result.current.activeSearch).toEqual({
			assetType: AssetType.SWITCHBOARD_PANEL,
			q: "PNL-001",
		});
	});

	it("treats an unknown type as all types", () => {
		const { result } = renderAt("/assets?q=PNL&type=GENERATOR");

		expect(result.current.activeSearch).toEqual({ assetType: null, q: "PNL" });
	});

	it("writes a search to the address and keeps other parameters", () => {
		const { result } = renderAt("/assets?panel=import");

		act(() => {
			result.current.submitSearch({
				assetType: AssetType.TRANSFORMER,
				q: "TX 1",
			});
		});

		const parameters = new URLSearchParams(result.current.search);
		expect(Object.fromEntries(parameters)).toEqual({
			panel: "import",
			q: "TX 1",
			type: "TRANSFORMER",
		});
	});

	it("leaves the type out for all types", () => {
		const { result } = renderAt("/assets?q=old&type=TRANSFORMER");

		act(() => {
			result.current.submitSearch({ assetType: null, q: "new" });
		});

		expect(result.current.search).toBe("?q=new");
	});

	it("clears the search and the type", () => {
		const { result } = renderAt("/assets?q=PNL&type=TRANSFORMER&panel=import");

		act(() => {
			result.current.clearSearch();
		});

		expect(result.current.search).toBe("?panel=import");
		expect(result.current.activeSearch).toBeNull();
	});

	it("adds a history entry, so Back returns to the previous search", () => {
		const { result } = renderAt("/assets?q=first");

		act(() => {
			result.current.submitSearch({ assetType: null, q: "second" });
		});
		act(() => {
			result.current.goBack();
		});

		expect(result.current.activeSearch).toEqual({
			assetType: null,
			q: "first",
		});
	});
});
