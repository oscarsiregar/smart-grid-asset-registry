import { render, screen, within } from "@testing-library/react";
import { Provider } from "react-redux";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { HTTPCode } from "~/libs/enums/enums.js";
import { config } from "~/libs/modules/config/config.js";
import { createStore } from "~/libs/modules/store/store.module.js";
import {
	type AssetSearchResultDto,
	AssetType,
	OperationalStatus,
} from "~/modules/assets/assets.js";

import { SearchResults } from "./search-results.js";

// Position of the last element for Array.prototype.at.
const LAST_POSITION = -1;

// In the order the API sends them; the component must not re-sort.
const RESULT: AssetSearchResultDto = {
	items: [
		{
			ancestorAssetIds: ["SS-001", "SWB-001-1"],
			assetId: "PNL-001-1-01",
			assetName: "Aurora Panel A1",
			assetType: AssetType.SWITCHBOARD_PANEL,
			childCount: 0,
			operationalStatus: OperationalStatus.MAINTENANCE,
		},
		{
			ancestorAssetIds: [],
			assetId: "SS-001",
			assetName: "Aurora Substation",
			assetType: AssetType.SUBSTATION,
			childCount: 5,
			operationalStatus: null,
		},
		{
			ancestorAssetIds: [],
			assetId: "..",
			assetName: "Dots Substation",
			assetType: AssetType.SUBSTATION,
			childCount: 0,
			operationalStatus: OperationalStatus.IN_SERVICE,
		},
	],
	totalItems: 3,
};

const stubSearch = (data: AssetSearchResultDto) => {
	const fetchMock = vi.fn(() =>
		Promise.resolve(
			Response.json(
				{ code: HTTPCode.OK, data, message: "ok", status: "success" },
				{ status: HTTPCode.OK },
			),
		),
	);
	vi.stubGlobal("fetch", fetchMock);

	return fetchMock;
};

const renderResults = (q = "aurora"): void => {
	render(
		<Provider store={createStore(config)}>
			<MemoryRouter>
				<SearchResults search={{ assetType: null, q }} />
			</MemoryRouter>
		</Provider>,
	);
};

afterEach(() => {
	vi.unstubAllGlobals();
});

describe("SearchResults", () => {
	it("is a labelled region with the number of matches", async () => {
		stubSearch(RESULT);
		renderResults();

		const region = await screen.findByRole("region", {
			name: "Search results",
		});
		expect(
			await within(region).findByText("3 assets match “aurora”"),
		).toBeInTheDocument();
	});

	it("lists each result as a link with ID, name, type, status and location", async () => {
		stubSearch(RESULT);
		renderResults();

		const panel = await screen.findByRole("link", { name: /PNL-001-1-01/ });
		expect(panel).toHaveAttribute("href", "/assets/PNL-001-1-01");
		expect(panel).toHaveTextContent("Aurora Panel A1");

		const [panelItem, substationItem] = screen.getAllByRole("listitem");

		if (!panelItem || !substationItem) {
			throw new Error("Expected the results to be listed");
		}

		expect(
			within(panelItem).getByText("Switchboard panel"),
		).toBeInTheDocument();
		expect(within(panelItem).getByText("Maintenance")).toBeInTheDocument();
		expect(
			within(panelItem).getByText("SS-001 › SWB-001-1"),
		).toBeInTheDocument();
		expect(
			within(substationItem).getByText("No status recorded"),
		).toBeInTheDocument();
		expect(within(substationItem).getByText("Top level")).toBeInTheDocument();
	});

	it("shows an asset with ID .. without a link", async () => {
		stubSearch(RESULT);
		renderResults();

		const items = await screen.findAllByRole("listitem");
		const dots = items.at(LAST_POSITION);

		if (!dots) {
			throw new Error("Expected the .. result");
		}

		expect(within(dots).queryByRole("link")).not.toBeInTheDocument();
		expect(within(dots).getByText("..")).toBeInTheDocument();
		expect(
			within(dots).getByText("This asset cannot be opened."),
		).toBeInTheDocument();
	});

	it("keeps the order received", async () => {
		stubSearch(RESULT);
		renderResults();

		await screen.findAllByRole("listitem");
		expect(
			screen.getAllByRole("listitem").map((item) => item.dataset["assetId"]),
		).toEqual(["PNL-001-1-01", "SS-001", ".."]);
	});

	it("sends the text it is given", async () => {
		const fetchMock = stubSearch(RESULT);
		renderResults("PNL 1");

		await screen.findAllByRole("listitem");
		const [[request]] = fetchMock.mock.calls as unknown as [[Request]];
		expect(new URL(request.url).searchParams.get("q")).toBe("PNL 1");
	});
});
