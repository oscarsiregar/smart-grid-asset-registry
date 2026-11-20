import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import {
	type AssetSearchRequestDto,
	AssetType,
} from "~/modules/assets/assets.js";

import { SearchForm } from "./search-form.js";

const WHITESPACE_ONLY = " \t ";

type RenderOptions = {
	activeSearch?: AssetSearchRequestDto | null;
};

const renderForm = ({ activeSearch = null }: RenderOptions = {}) => {
	const onSubmit = vi.fn();
	const onClear = vi.fn();
	render(
		<SearchForm
			activeSearch={activeSearch}
			onClear={onClear}
			onSubmit={onSubmit}
		/>,
	);

	return { onClear, onSubmit };
};

const getField = (): HTMLElement =>
	screen.getByRole("textbox", { name: "Search assets" });

describe("SearchForm", () => {
	it("is a search landmark with a labelled field", () => {
		renderForm();

		expect(screen.getByRole("search")).toBeInTheDocument();
		expect(getField()).toHaveValue("");
	});

	it("keeps the labels for assistive technology only and shows the placeholder", () => {
		renderForm();

		expect(getField()).toHaveAttribute("placeholder", "Search by ID or name");
		expect(screen.getByText("Search assets")).toHaveClass("visually-hidden");
		expect(screen.getByText("Asset type")).toHaveClass("visually-hidden");
		expect(
			screen.getByRole("combobox", { name: "Asset type" }),
		).toBeInTheDocument();
	});

	it("submits the trimmed text with the Search button", async () => {
		const { onSubmit } = renderForm();

		fireEvent.change(getField(), { target: { value: "  PNL-001  " } });
		fireEvent.click(screen.getByRole("button", { name: "Search" }));

		await waitFor(() => {
			expect(onSubmit).toHaveBeenCalledWith({ assetType: null, q: "PNL-001" });
		});
	});

	it("submits with Enter in the field", async () => {
		const { onSubmit } = renderForm();

		fireEvent.change(getField(), { target: { value: "TX" } });
		fireEvent.submit(getField());

		await waitFor(() => {
			expect(onSubmit).toHaveBeenCalledWith({ assetType: null, q: "TX" });
		});
	});

	it("refuses a blank text with a message and does not submit", async () => {
		const { onSubmit } = renderForm();

		fireEvent.change(getField(), { target: { value: WHITESPACE_ONLY } });
		fireEvent.click(screen.getByRole("button", { name: "Search" }));

		expect(
			await screen.findByText("Enter an asset ID or name to search."),
		).toBeInTheDocument();
		expect(onSubmit).not.toHaveBeenCalled();
	});

	it("starts with the search from the address and keeps its type", async () => {
		const { onSubmit } = renderForm({
			activeSearch: { assetType: AssetType.TRANSFORMER, q: "TX-001" },
		});

		expect(getField()).toHaveValue("TX-001");

		fireEvent.click(screen.getByRole("button", { name: "Search" }));

		await waitFor(() => {
			expect(onSubmit).toHaveBeenCalledWith({
				assetType: AssetType.TRANSFORMER,
				q: "TX-001",
			});
		});
	});

	it("offers Clear only while a search is shown", () => {
		renderForm();
		expect(
			screen.queryByRole("button", { name: "Clear" }),
		).not.toBeInTheDocument();
	});

	it("clears the search", () => {
		const { onClear } = renderForm({
			activeSearch: { assetType: null, q: "PNL" },
		});

		fireEvent.click(screen.getByRole("button", { name: "Clear" }));

		expect(onClear).toHaveBeenCalledOnce();
	});

	it("submits the chosen type", async () => {
		const { onSubmit } = renderForm();

		fireEvent.change(getField(), { target: { value: "001" } });
		fireEvent.change(screen.getByRole("combobox", { name: "Asset type" }), {
			target: { value: AssetType.SWITCHBOARD_PANEL },
		});
		fireEvent.click(screen.getByRole("button", { name: "Search" }));

		await waitFor(() => {
			expect(onSubmit).toHaveBeenCalledWith({
				assetType: AssetType.SWITCHBOARD_PANEL,
				q: "001",
			});
		});
	});

	it("starts with the type from the address", () => {
		renderForm({
			activeSearch: { assetType: AssetType.TRANSFORMER, q: "TX" },
		});

		expect(
			screen.getByRole("combobox", { name: "Asset type" }),
		).toHaveDisplayValue("Transformer");
	});
});
