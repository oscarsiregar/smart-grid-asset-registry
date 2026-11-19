import { fireEvent, render, screen } from "@testing-library/react";
import { Provider } from "react-redux";
import { MemoryRouter, Route, Routes, useLocation } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { AppRoute } from "~/libs/enums/enums.js";
import { config } from "~/libs/modules/config/config.js";
import { createStore } from "~/libs/modules/store/store.module.js";

import { Assets } from "./assets.js";

const LocationProbe: React.FC = () => {
	const { search } = useLocation();

	return <output data-testid="search">{search}</output>;
};

const renderAssets = (initialEntry: string): void => {
	render(
		<Provider store={createStore(config)}>
			<MemoryRouter initialEntries={[initialEntry]}>
				<Routes>
					<Route
						element={
							<>
								<Assets />
								<LocationProbe />
							</>
						}
						path={AppRoute.ASSETS}
					/>
				</Routes>
			</MemoryRouter>
		</Provider>,
	);
};

const getDialog = (): HTMLElement =>
	screen.getByRole("dialog", { hidden: true });

describe("Assets page", () => {
	it("shows the empty state with the panel closed", () => {
		renderAssets(AppRoute.ASSETS);

		expect(
			screen.getByText("No assets have been imported yet."),
		).toBeInTheDocument();
		expect(getDialog()).not.toHaveAttribute("open");
	});

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
