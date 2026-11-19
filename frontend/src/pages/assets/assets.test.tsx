import { fireEvent, render, screen } from "@testing-library/react";
import { Provider } from "react-redux";
import { MemoryRouter, Route, Routes, useLocation } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { AppRoute, HTTPCode } from "~/libs/enums/enums.js";
import { config } from "~/libs/modules/config/config.js";
import { createStore } from "~/libs/modules/store/store.module.js";
import {
	type AssetListDto,
	AssetType,
	OperationalStatus,
} from "~/modules/assets/assets.js";

import { Assets } from "./assets.js";

const EMPTY: AssetListDto = { items: [], totalItems: 0 };

const TOP_LEVEL: AssetListDto = {
	items: [
		{
			assetId: "SS-001",
			assetName: "Aurora Substation",
			assetType: AssetType.SUBSTATION,
			childCount: 5,
			operationalStatus: OperationalStatus.IN_SERVICE,
		},
	],
	totalItems: 1,
};

const respondWith = (status: number, data: unknown): Promise<Response> =>
	Promise.resolve(
		Response.json(
			status === HTTPCode.OK
				? { code: status, data, message: "ok", status: "success" }
				: {
						code: status,
						data: null,
						message: "Internal Server Error",
						status: "error",
					},
			{ status },
		),
	);

const stubTopLevel = (status: number, data: unknown): void => {
	vi.stubGlobal(
		"fetch",
		vi.fn(() => respondWith(status, data)),
	);
};

const LocationProbe: React.FC = () => {
	const { pathname, search } = useLocation();

	return (
		<>
			<output data-testid="pathname">{pathname}</output>
			<output data-testid="search">{search}</output>
		</>
	);
};

const page = (
	<>
		<Assets />
		<LocationProbe />
	</>
);

const renderAssets = (initialEntry: string): void => {
	render(
		<Provider store={createStore(config)}>
			<MemoryRouter initialEntries={[initialEntry]}>
				<Routes>
					<Route element={page} path={AppRoute.ASSETS} />
					<Route element={page} path={AppRoute.ASSETS_$ASSET_ID} />
				</Routes>
			</MemoryRouter>
		</Provider>,
	);
};

const getDialog = (): HTMLElement =>
	screen.getByRole("dialog", { hidden: true });

beforeEach(() => {
	stubTopLevel(HTTPCode.OK, EMPTY);
});

afterEach(() => {
	vi.unstubAllGlobals();
});

describe("Assets page: states", () => {
	it("says the assets are loading in the tree area", () => {
		vi.stubGlobal(
			"fetch",
			vi.fn(
				() =>
					new Promise<Response>(() => {
						// Never settles.
					}),
			),
		);

		renderAssets(AppRoute.ASSETS);

		expect(screen.getByText("Loading assets…")).toBeInTheDocument();
	});

	it("shows the empty registry without a tree; Import CSV opens the panel", async () => {
		renderAssets(AppRoute.ASSETS);

		expect(
			await screen.findByText("No assets have been imported yet."),
		).toBeInTheDocument();
		expect(screen.queryByRole("tree")).not.toBeInTheDocument();
		expect(getDialog()).not.toHaveAttribute("open");

		fireEvent.click(screen.getByRole("button", { name: "Import CSV" }));

		expect(getDialog()).toHaveAttribute("open");
	});

	it("shows a top-level failure with a retry, and Import CSV still works", async () => {
		stubTopLevel(HTTPCode.INTERNAL_SERVER_ERROR, null);

		renderAssets(AppRoute.ASSETS);

		expect(
			await screen.findByText("The assets could not be loaded."),
		).toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: "Try again" }),
		).toBeInTheDocument();

		fireEvent.click(screen.getByRole("button", { name: "Import CSV" }));

		expect(getDialog()).toHaveAttribute("open");
	});

	it("shows the tree beside the details and selects through the address", async () => {
		stubTopLevel(HTTPCode.OK, TOP_LEVEL);

		renderAssets(AppRoute.ASSETS);

		expect(await screen.findByRole("tree")).toBeInTheDocument();
		expect(
			screen.getByText("Select an asset in the tree to see its details."),
		).toBeInTheDocument();

		fireEvent.click(screen.getByRole("link", { name: /SS-001/ }));

		expect(screen.getByTestId("pathname")).toHaveTextContent("/assets/SS-001");
	});
});

describe("Assets page: import panel", () => {
	it("opens the import panel and records it in the URL", () => {
		renderAssets(AppRoute.ASSETS);

		fireEvent.click(screen.getByRole("button", { name: "Import CSV" }));

		expect(getDialog()).toHaveAttribute("open");
		expect(screen.getByTestId("search")).toHaveTextContent("?panel=import");
	});

	it("opens the import panel from a deep link", () => {
		renderAssets(`${AppRoute.ASSETS}?panel=import`);

		expect(getDialog()).toHaveAttribute("open");
		expect(
			screen.getByRole("heading", { name: "Import assets from CSV" }),
		).toBeInTheDocument();
	});

	it("closes the panel on Escape and removes it from the URL", () => {
		renderAssets(`${AppRoute.ASSETS}?panel=import`);

		fireEvent(getDialog(), new Event("cancel", { cancelable: true }));

		expect(getDialog()).not.toHaveAttribute("open");
		expect(screen.getByTestId("search")).toBeEmptyDOMElement();
	});

	it("shows the empty import form on a deep link (results are not kept)", () => {
		renderAssets(`${AppRoute.ASSETS}?panel=import`);

		expect(screen.getByLabelText("CSV file")).toBeInTheDocument();
		expect(screen.getByRole("button", { name: "Check file" })).toBeDisabled();
	});

	it("closes the panel and removes it from the URL", () => {
		renderAssets(`${AppRoute.ASSETS}?panel=import`);

		fireEvent.click(screen.getByRole("button", { name: "Close" }));

		expect(getDialog()).not.toHaveAttribute("open");
		expect(screen.getByTestId("search")).toBeEmptyDOMElement();
	});
});
