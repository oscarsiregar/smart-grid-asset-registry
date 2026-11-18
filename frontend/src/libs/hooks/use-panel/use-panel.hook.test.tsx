import { act, renderHook } from "@testing-library/react";
import { MemoryRouter, useLocation, useNavigate } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { PanelName } from "~/libs/enums/enums.js";

import { usePanel } from "./use-panel.hook.js";

const BACK = -1;

const renderPanelHook = (initialEntry: string) => {
	const wrapper: React.FC<{ children: React.ReactNode }> = ({ children }) => (
		<MemoryRouter initialEntries={[initialEntry]}>{children}</MemoryRouter>
	);

	return renderHook(
		() => ({
			location: useLocation(),
			navigate: useNavigate(),
			panel: usePanel(PanelName.IMPORT),
		}),
		{ wrapper },
	);
};

describe("usePanel", () => {
	it("is closed without the search parameter", () => {
		const { result } = renderPanelHook("/assets");

		expect(result.current.panel.isOpen).toBe(false);
	});

	it("is open on a deep link", () => {
		const { result } = renderPanelHook("/assets?panel=import");

		expect(result.current.panel.isOpen).toBe(true);
	});

	it("ignores an unknown panel name", () => {
		const { result } = renderPanelHook("/assets?panel=unknown");

		expect(result.current.panel.isOpen).toBe(false);
	});

	it("opens by adding the parameter and keeps other parameters", () => {
		const { result } = renderPanelHook("/assets?asset=SS-001");

		act(() => {
			result.current.panel.openPanel();
		});

		expect(result.current.panel.isOpen).toBe(true);
		expect(result.current.location.search).toBe("?asset=SS-001&panel=import");
	});

	it("closes by removing the parameter", () => {
		const { result } = renderPanelHook("/assets?panel=import");

		act(() => {
			result.current.panel.closePanel();
		});

		expect(result.current.panel.isOpen).toBe(false);
		expect(result.current.location.search).toBe("");
	});

	it("closes on browser Back after opening", async () => {
		const { result } = renderPanelHook("/assets");

		act(() => {
			result.current.panel.openPanel();
		});
		await act(async () => {
			await result.current.navigate(BACK);
		});

		expect(result.current.panel.isOpen).toBe(false);
	});
});
