import {
	fireEvent,
	render,
	screen,
	waitFor,
	within,
} from "@testing-library/react";
import { Provider } from "react-redux";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { HTTPCode } from "~/libs/enums/enums.js";
import { config } from "~/libs/modules/config/config.js";
import { bindShowNotification } from "~/libs/modules/notification/notification.js";
import { createStore } from "~/libs/modules/store/store.module.js";
import {
	type AssetSearchRequestDto,
	type AssetSearchResultDto,
	AssetType,
	OperationalStatus,
} from "~/modules/assets/assets.js";

import { SearchResults } from "./search-results.js";

// Placeholder result rows while a search runs (spec 005).
const SEARCH_SKELETON_ROWS = 4;

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

const renderResults = (
	q = "aurora",
	assetType: AssetSearchRequestDto["assetType"] = null,
): void => {
	render(
		<Provider store={createStore(config)}>
			<MemoryRouter>
				<SearchResults search={{ assetType, q }} />
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

	it("names the type in the summary and sends it", async () => {
		const fetchMock = stubSearch(RESULT);
		renderResults("aurora", AssetType.SWITCHBOARD_PANEL);

		expect(
			await screen.findByText(
				"3 assets match “aurora” of type Switchboard panel",
			),
		).toBeInTheDocument();
		const [[request]] = fetchMock.mock.calls as unknown as [[Request]];
		expect(new URL(request.url).searchParams.get("assetType")).toBe(
			"SWITCHBOARD_PANEL",
		);
	});

	it("says it is searching while the request runs", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn(() => new Promise<Response>(() => {})),
		);
		renderResults();

		expect(await screen.findByRole("status")).toHaveTextContent("Searching…");
		expect(document.querySelectorAll("[data-skeleton-row]")).toHaveLength(
			SEARCH_SKELETON_ROWS,
		);
	});
});

// ---- The states of a search (user story 3) ---------------------------------

// The first request and the retry.
const RETRIED_REQUESTS = 2;

// A broad search: more matches than the backend's first page.
const BROAD_TOTAL = 1234;

const stubError = (status: number, body: unknown) => {
	const fetchMock = vi.fn(() =>
		Promise.resolve(Response.json(body, { status })),
	);
	vi.stubGlobal("fetch", fetchMock);

	return fetchMock;
};

const TOO_LONG_MESSAGE = "Invalid request: q must be at most 100 characters";

const stubRefused = () =>
	stubError(HTTPCode.BAD_REQUEST, {
		code: HTTPCode.BAD_REQUEST,
		data: { errorCode: "INVALID_REQUEST" },
		message: TOO_LONG_MESSAGE,
		status: "error",
	});

const stubFailure = () =>
	stubError(HTTPCode.INTERNAL_SERVER_ERROR, {
		code: HTTPCode.INTERNAL_SERVER_ERROR,
		data: null,
		message: "Internal Server Error",
		status: "error",
	});

describe("SearchResults states", () => {
	it("says no asset matches, with a hint, and lists nothing", async () => {
		stubSearch({ items: [], totalItems: 0 });
		const showNotificationMock = vi.fn();
		const unbind = bindShowNotification(showNotificationMock);
		renderResults("NO-SUCH-ASSET");

		expect(
			await screen.findByText("No asset matches “NO-SUCH-ASSET”"),
		).toBeInTheDocument();
		expect(
			screen.getByText("Check the spelling, or search all types."),
		).toBeInTheDocument();
		expect(screen.queryByRole("list")).not.toBeInTheDocument();
		expect(showNotificationMock).not.toHaveBeenCalled();
		unbind();
	});

	it("names the type when nothing of that type matches", async () => {
		stubSearch({ items: [], totalItems: 0 });
		renderResults("aurora", AssetType.TRANSFORMER);

		expect(
			await screen.findByText("No asset matches “aurora” of type Transformer"),
		).toBeInTheDocument();
	});

	it("says how many of the matches are shown when there are more", async () => {
		stubSearch({ items: RESULT.items, totalItems: BROAD_TOTAL });
		renderResults("-");

		// The count shown is the number of items received.
		expect(
			await screen.findByText(
				`Showing ${String(RESULT.items.length)} of ${String(BROAD_TOTAL)} assets that match “-”. Add more of the ID or name, or choose a type, to narrow the search.`,
			),
		).toBeInTheDocument();
		expect(screen.getAllByRole("listitem")).toHaveLength(RESULT.items.length);
	});

	it("shows a refused search's message as sent, without a retry or a toast", async () => {
		stubRefused();
		const showNotificationMock = vi.fn();
		const unbind = bindShowNotification(showNotificationMock);
		renderResults("x");

		expect(await screen.findByText(TOO_LONG_MESSAGE)).toBeInTheDocument();
		expect(
			screen.queryByRole("button", { name: "Try again" }),
		).not.toBeInTheDocument();
		expect(
			screen.queryByText("The search could not be completed."),
		).not.toBeInTheDocument();
		expect(showNotificationMock).not.toHaveBeenCalled();
		unbind();
	});

	it("says a failed search could not be completed and retries it", async () => {
		const fetchMock = stubFailure();
		renderResults();

		expect(
			await screen.findByText("The search could not be completed."),
		).toBeInTheDocument();
		expect(fetchMock).toHaveBeenCalledOnce();

		fireEvent.click(screen.getByRole("button", { name: "Try again" }));

		await waitFor(() => {
			expect(fetchMock).toHaveBeenCalledTimes(RETRIED_REQUESTS);
		});
	});
});
