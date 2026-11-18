import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes, useLocation } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { AppRoute } from "~/libs/enums/enums.js";

import { Assets } from "./assets.js";

const LocationProbe: React.FC = () => {
	const { search } = useLocation();

	return <output data-testid="search">{search}</output>;
};

const renderAssets = (initialEntry: string): void => {
	render(
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
		</MemoryRouter>,
	);
};

const getDialog = (): HTMLElement =>
	screen.getByRole("dialog", { hidden: true });

describe("Assets page", () => {
	it("shows the empty state with the panel closed", () => {
		renderAssets(AppRoute.ASSETS);

		expect(
			screen.getByText("No assets yet. Import a CSV file to get started."),
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

	it("closes the panel and removes it from the URL", () => {
		renderAssets(`${AppRoute.ASSETS}?panel=import`);

		fireEvent.click(screen.getByRole("button", { name: "Close" }));

		expect(getDialog()).not.toHaveAttribute("open");
		expect(screen.getByTestId("search")).toBeEmptyDOMElement();
	});
});
