import { fireEvent, render, screen, within } from "@testing-library/react";
import { useCallback } from "react";
import { Provider } from "react-redux";
import {
	MemoryRouter,
	Route,
	Routes,
	useLocation,
	useNavigate,
} from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { AppRoute, HTTPCode } from "~/libs/enums/enums.js";
import { config } from "~/libs/modules/config/config.js";
import { createStore } from "~/libs/modules/store/store.module.js";
import {
	type AssetDetailDto,
	type AssetListDto,
	type AssetNodeDto,
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

// History steps for the Back and Forward test buttons.
const HISTORY_BACK = -1;
const HISTORY_FORWARD = 1;

const HistoryControls: React.FC = () => {
	const navigate = useNavigate();

	const handleBackClick = useCallback((): void => {
		void navigate(HISTORY_BACK);
	}, [navigate]);

	const handleForwardClick = useCallback((): void => {
		void navigate(HISTORY_FORWARD);
	}, [navigate]);

	return (
		<>
			<button onClick={handleBackClick} type="button">
				History back
			</button>
			<button onClick={handleForwardClick} type="button">
				History forward
			</button>
		</>
	);
};

const page = (
	<>
		<Assets />
		<LocationProbe />
		<HistoryControls />
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
		// No placeholder stays after a failure (spec 005 SC-004).
		expect(document.querySelector("[data-shape]")).toBeNull();

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

	it("has the Asset explorer header and the register card in the sidebar", async () => {
		stubTopLevel(HTTPCode.OK, TOP_LEVEL);

		renderAssets(AppRoute.ASSETS);

		expect(
			screen.getByRole("heading", { level: 1, name: "Asset explorer" }),
		).toBeInTheDocument();
		const sidebar = screen.getByRole("complementary", {
			name: "Asset navigation",
		});
		expect(await within(sidebar).findByText("substation")).toHaveTextContent(
			"1 substation",
		);
		expect(within(sidebar).getByText("Register")).toBeInTheDocument();
	});

	it("shows no register card for an empty registry", async () => {
		renderAssets(AppRoute.ASSETS);

		expect(
			await screen.findByText("No assets have been imported yet."),
		).toBeInTheDocument();
		expect(screen.queryByText("Register")).not.toBeInTheDocument();
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

// ---- Addresses (user story 3) -----------------------------------------------

const createNode = ({
	assetId,
	assetType,
	childCount,
}: Pick<
	AssetNodeDto,
	"assetId" | "assetType" | "childCount"
>): AssetNodeDto => ({
	assetId,
	assetName: `Name ${assetId}`,
	assetType,
	childCount,
	operationalStatus: OperationalStatus.MAINTENANCE,
});

const SUBSTATION = createNode({
	assetId: "SS-001",
	assetType: AssetType.SUBSTATION,
	childCount: 1,
});
const DOTS = createNode({
	assetId: "..",
	assetType: AssetType.SUBSTATION,
	childCount: 1,
});
const SWITCHBOARD = createNode({
	assetId: "SWB-001-1",
	assetType: AssetType.SWITCHBOARD,
	childCount: 2,
});
const PANEL_A = createNode({
	assetId: "PNL-001-1-01",
	assetType: AssetType.SWITCHBOARD_PANEL,
	childCount: 0,
});
const PANEL_B = createNode({
	assetId: "PNL-001-1-02",
	assetType: AssetType.SWITCHBOARD_PANEL,
	childCount: 0,
});
const DOTS_CHILD = createNode({
	assetId: "TX-DOT-1",
	assetType: AssetType.TRANSFORMER,
	childCount: 0,
});
const SLASH = createNode({
	assetId: "SS/9 A",
	assetType: AssetType.SUBSTATION,
	childCount: 0,
});

const createList = (items: AssetNodeDto[]): AssetListDto => ({
	items,
	totalItems: items.length,
});

// The last step of the path is the asset itself.
const createDetail = (path: AssetNodeDto[]): AssetDetailDto | null => {
	const asset = path.at(HISTORY_BACK);

	return asset
		? {
				asset: {
					assetId: asset.assetId,
					assetName: asset.assetName,
					assetType: asset.assetType,
					commissionedDate: null,
					manufacturer: null,
					model: null,
					operationalStatus: asset.operationalStatus,
					parentAssetId: null,
					ratingKva: null,
					serialNumber: null,
					voltageKv: null,
				},
				children: [],
				descendantCounts: null,
				path,
			}
		: null;
};

const API = "http://localhost/api/assets";

const ROUTED_RESPONSES = new Map<string, unknown>([
	[`${API}/PNL-001-1-01`, createDetail([SUBSTATION, SWITCHBOARD, PANEL_A])],
	[`${API}/PNL-001-1-02`, createDetail([SUBSTATION, SWITCHBOARD, PANEL_B])],
	[`${API}/SS%2F9%20A`, createDetail([SLASH])],
	[`${API}/SS-001`, createDetail([SUBSTATION])],
	[`${API}?parentAssetId=..`, createList([DOTS_CHILD])],
	[`${API}?parentAssetId=SS-001`, createList([SWITCHBOARD])],
	[`${API}?parentAssetId=SWB-001-1`, createList([PANEL_A, PANEL_B])],
	[
		`${API}?q=PNL-001-1-01`,
		{
			items: [{ ...PANEL_A, ancestorAssetIds: ["SS-001", "SWB-001-1"] }],
			totalItems: 1,
		},
	],
	[API, createList([SUBSTATION, DOTS])],
]);

// Answers by URL; anything unknown is the API's ASSET_NOT_FOUND.
const stubRoutedApi = () => {
	const fetchMock = vi.fn((request: Request): Promise<Response> => {
		const data = ROUTED_RESPONSES.get(request.url);

		return data === undefined
			? Promise.resolve(
					Response.json(
						{
							code: HTTPCode.NOT_FOUND,
							data: { errorCode: "ASSET_NOT_FOUND" },
							message: "No asset with ID NOPE exists.",
							status: "error",
						},
						{ status: HTTPCode.NOT_FOUND },
					),
				)
			: respondWith(HTTPCode.OK, data);
	});
	vi.stubGlobal("fetch", fetchMock);

	return fetchMock;
};

// A node's accessible name includes its descendants, so look it up by ID;
// waits until its parent's children have loaded.
const findNode = async (assetId: string): Promise<HTMLElement> =>
	await vi.waitFor(() => {
		const node = screen
			.queryAllByRole("treeitem")
			.find((element) => element.dataset["assetId"] === assetId);

		if (!node) {
			throw new Error(`No tree node ${assetId}`);
		}

		return node;
	});

const getSelectedNodes = (): HTMLElement[] =>
	screen
		.queryAllByRole("treeitem")
		.filter((element) => element.getAttribute("aria-selected") === "true");

const findSelectedNode = async (assetId: string): Promise<HTMLElement> => {
	const node = await findNode(assetId);
	await vi.waitFor(() => {
		expect(node).toHaveAttribute("aria-selected", "true");
	});

	return node;
};

describe("Assets page: addresses", () => {
	it("opens a deep link with the ancestors expanded and the asset highlighted and scrolled into view", async () => {
		stubRoutedApi();
		const scrollSpy = vi.spyOn(Element.prototype, "scrollIntoView");

		renderAssets("/assets/PNL-001-1-01");

		const panel = await findSelectedNode("PNL-001-1-01");
		expect(await findNode("SS-001")).toHaveAttribute("aria-expanded", "true");
		expect(await findNode("SWB-001-1")).toHaveAttribute(
			"aria-expanded",
			"true",
		);
		expect(within(panel).getByText("Selected")).toBeInTheDocument();
		expect(getSelectedNodes()).toEqual([panel]);
		expect(scrollSpy.mock.contexts).toContain(panel);
		scrollSpy.mockRestore();
	});

	it("moves between selections with back and forward", async () => {
		stubRoutedApi();
		renderAssets("/assets/PNL-001-1-01");
		await findSelectedNode("PNL-001-1-01");

		fireEvent.click(within(await findNode("PNL-001-1-02")).getByRole("link"));
		await findSelectedNode("PNL-001-1-02");

		fireEvent.click(screen.getByRole("button", { name: "History back" }));
		expect(screen.getByTestId("pathname")).toHaveTextContent(
			"/assets/PNL-001-1-01",
		);
		await findSelectedNode("PNL-001-1-01");

		fireEvent.click(screen.getByRole("button", { name: "History forward" }));
		await findSelectedNode("PNL-001-1-02");
	});

	it("decodes an encoded address and requests that asset", async () => {
		const fetchMock = stubRoutedApi();

		renderAssets("/assets/SS%2F9%20A");

		expect(
			await screen.findByRole("heading", { name: "Name SS/9 A" }),
		).toBeInTheDocument();
		expect(fetchMock.mock.calls.map(([request]) => request.url)).toContain(
			`${API}/SS%2F9%20A`,
		);
	});

	it.each(["/assets", "/assets/"])("selects nothing at %s", async (address) => {
		stubRoutedApi();

		renderAssets(address);

		await findNode("SS-001");
		expect(
			screen.getByText("Select an asset in the tree to see its details."),
		).toBeInTheDocument();
		expect(getSelectedNodes()).toEqual([]);
	});

	it("lists an asset with ID .. and lets it expand, but not be selected", async () => {
		stubRoutedApi();
		renderAssets("/assets");

		const dots = await findNode("..");
		expect(within(dots).getByText("Name ..")).toBeInTheDocument();
		expect(within(dots).getByText("Maintenance")).toBeInTheDocument();
		expect(within(dots).queryByRole("link")).not.toBeInTheDocument();

		fireEvent.click(screen.getByRole("button", { name: "Expand .." }));

		expect(await screen.findByText("TX-DOT-1")).toBeInTheDocument();
	});

	it("shows an unknown ID as not found while the tree keeps working", async () => {
		stubRoutedApi();

		renderAssets("/assets/NOPE");

		expect(
			await screen.findByText("No asset with ID NOPE exists."),
		).toBeInTheDocument();
		expect(screen.getByRole("tree")).toBeInTheDocument();
		expect(
			screen.queryByText("The asset could not be loaded."),
		).not.toBeInTheDocument();
	});
});

// ---- The import panel on the explorer (user story 4) ------------------------

const IMPORT_COUNTS = {
	imported: 1,
	rejected: 0,
	total: 1,
	unchanged: 0,
	warningRows: 0,
};

const createImportResult = (isCommitted: boolean) => ({
	counts: IMPORT_COUNTS,
	fileName: "grid_assets.csv",
	import: isCommitted
		? { createdAt: "2026-11-19T10:15:00Z", id: "6f1c2c1e" }
		: null,
	isCommitted,
	mode: "PARTIAL",
	outcome: isCommitted ? "COMMITTED" : "VALIDATED_ONLY",
	rejectedRows: [],
	warnings: [],
});

// An empty registry until the import commits; then one substation.
const serveFirstImport = (): void => {
	let isCommitted = false;
	vi.stubGlobal(
		"fetch",
		vi.fn((request: Request): Promise<Response> => {
			if (request.url.endsWith("/imports/validate")) {
				return respondWith(HTTPCode.OK, createImportResult(false));
			}

			if (request.url.endsWith("/imports")) {
				isCommitted = true;

				return respondWith(HTTPCode.OK, createImportResult(true));
			}

			return respondWith(HTTPCode.OK, isCommitted ? TOP_LEVEL : EMPTY);
		}),
	);
};

describe("Assets page: import panel on the explorer", () => {
	it("keeps the selected asset in the address while the panel opens and closes", async () => {
		stubRoutedApi();
		renderAssets("/assets/SS-001");
		await findSelectedNode("SS-001");

		fireEvent.click(screen.getByRole("button", { name: "Import CSV" }));

		expect(getDialog()).toHaveAttribute("open");
		expect(screen.getByTestId("pathname")).toHaveTextContent("/assets/SS-001");
		expect(screen.getByTestId("search")).toHaveTextContent("?panel=import");

		fireEvent.click(screen.getByRole("button", { name: "Close" }));

		expect(getDialog()).not.toHaveAttribute("open");
		expect(screen.getByTestId("pathname")).toHaveTextContent("/assets/SS-001");
		expect(screen.getByTestId("search")).toBeEmptyDOMElement();
		expect(await findSelectedNode("SS-001")).toBeInTheDocument();
	});

	it("replaces the empty state with the tree after the first committed import", async () => {
		serveFirstImport();
		renderAssets("/assets");
		await screen.findByText("No assets have been imported yet.");

		fireEvent.click(screen.getByRole("button", { name: "Import CSV" }));
		fireEvent.change(screen.getByLabelText("CSV file"), {
			target: { files: [new File(["asset_id"], "grid_assets.csv")] },
		});
		fireEvent.click(screen.getByRole("button", { name: "Check file" }));
		fireEvent.click(
			await screen.findByRole("button", { name: "Save 1 assets" }),
		);
		await screen.findByText(/^Committed/);
		fireEvent.click(screen.getByRole("button", { name: "Close" }));

		expect(await screen.findByRole("tree")).toBeInTheDocument();
		expect(
			screen.queryByText("No assets have been imported yet."),
		).not.toBeInTheDocument();
		expect(await findNode("SS-001")).toBeInTheDocument();
	});
});

// ---- The search on the explorer (feature 003, user story 1) ----------------

describe("Assets page: search", () => {
	it("shows the results in place of the tree, the details as without a search", async () => {
		stubRoutedApi();
		renderAssets("/assets?q=PNL-001-1-01");

		expect(
			await screen.findByRole("region", { name: "Search results" }),
		).toBeInTheDocument();
		expect(
			await screen.findByRole("link", { name: /PNL-001-1-01/ }),
		).toBeInTheDocument();
		expect(screen.queryByRole("tree")).not.toBeInTheDocument();
		expect(
			screen.getByText("Select an asset in the tree to see its details."),
		).toBeInTheDocument();
	});

	it("opens a result through feature 002: the address, the reveal and the highlight", async () => {
		stubRoutedApi();
		renderAssets("/assets?q=PNL-001-1-01");

		fireEvent.click(await screen.findByRole("link", { name: /PNL-001-1-01/ }));

		expect(screen.getByTestId("pathname")).toHaveTextContent(
			"/assets/PNL-001-1-01",
		);
		expect(screen.getByTestId("search")).toBeEmptyDOMElement();
		await findSelectedNode("PNL-001-1-01");
		expect(await findNode("SS-001")).toHaveAttribute("aria-expanded", "true");
		expect(await findNode("SWB-001-1")).toHaveAttribute(
			"aria-expanded",
			"true",
		);
	});

	it("returns to the search and its results with Back", async () => {
		stubRoutedApi();
		renderAssets("/assets?q=PNL-001-1-01");
		fireEvent.click(await screen.findByRole("link", { name: /PNL-001-1-01/ }));
		await findSelectedNode("PNL-001-1-01");

		fireEvent.click(screen.getByRole("button", { name: "History back" }));

		expect(screen.getByTestId("search")).toHaveTextContent("?q=PNL-001-1-01");
		expect(
			await screen.findByRole("region", { name: "Search results" }),
		).toBeInTheDocument();
		expect(screen.getByRole("textbox", { name: "Search assets" })).toHaveValue(
			"PNL-001-1-01",
		);
	});

	it("writes a submitted search to the address", async () => {
		stubRoutedApi();
		renderAssets("/assets");
		await findNode("SS-001");

		fireEvent.change(screen.getByRole("textbox", { name: "Search assets" }), {
			target: { value: "PNL-001-1-01" },
		});
		fireEvent.click(screen.getByRole("button", { name: "Search" }));

		expect(
			await screen.findByRole("region", { name: "Search results" }),
		).toBeInTheDocument();
		expect(screen.getByTestId("search")).toHaveTextContent("?q=PNL-001-1-01");
	});

	it("brings back the tree and removes the search with Clear", async () => {
		stubRoutedApi();
		renderAssets("/assets?q=PNL-001-1-01");
		await screen.findByRole("region", { name: "Search results" });

		fireEvent.click(screen.getByRole("button", { name: "Clear" }));

		expect(await screen.findByRole("tree")).toBeInTheDocument();
		expect(screen.getByTestId("search")).toBeEmptyDOMElement();
	});

	it("offers no search on an empty registry", async () => {
		renderAssets("/assets");

		expect(
			await screen.findByText("No assets have been imported yet."),
		).toBeInTheDocument();
		expect(screen.queryByRole("search")).not.toBeInTheDocument();
	});
});

// ---- A search's states on the page (feature 003, user story 3) ------------

type SearchAnswer = "failed" | "pending";

// The routed API, except that every search fails or never answers.
const stubSearchAnswer = (answer: SearchAnswer): void => {
	const routed = stubRoutedApi();
	vi.stubGlobal(
		"fetch",
		vi.fn((request: Request): Promise<Response> => {
			if (!request.url.startsWith(`${API}?q=`)) {
				return routed(request);
			}

			return answer === "failed"
				? respondWith(HTTPCode.INTERNAL_SERVER_ERROR, null)
				: new Promise<Response>(() => {});
		}),
	);
};

describe("Assets page: search states", () => {
	it("keeps the details area working when the search fails; Clear brings the tree back", async () => {
		stubSearchAnswer("failed");
		renderAssets("/assets/SS-001?q=x");

		expect(
			await screen.findByText("The search could not be completed."),
		).toBeInTheDocument();
		expect(
			await screen.findByRole("heading", { name: SUBSTATION.assetName }),
		).toBeInTheDocument();
		// No placeholder stays after a failure (spec 005 SC-004).
		expect(document.querySelector("[data-skeleton-row]")).toBeNull();

		fireEvent.click(screen.getByRole("button", { name: "Clear" }));

		expect(await screen.findByRole("tree")).toBeInTheDocument();
		expect(screen.getByTestId("pathname")).toHaveTextContent("/assets/SS-001");
		expect(screen.getByTestId("search")).toBeEmptyDOMElement();
	});

	it("keeps the search when the import panel opens and closes", async () => {
		stubRoutedApi();
		renderAssets("/assets?q=PNL-001-1-01");
		await screen.findByRole("link", { name: /PNL-001-1-01/ });

		fireEvent.click(screen.getByRole("button", { name: "Import CSV" }));

		expect(getDialog()).toHaveAttribute("open");
		const opened = new URLSearchParams(
			screen.getByTestId("search").textContent,
		);
		expect(opened.get("q")).toBe("PNL-001-1-01");
		expect(opened.get("panel")).toBe("import");

		fireEvent.click(screen.getByRole("button", { name: "Close" }));

		expect(screen.getByTestId("search")).toHaveTextContent("?q=PNL-001-1-01");
		expect(
			screen.getByRole("region", { name: "Search results" }),
		).toBeInTheDocument();
	});

	it("lets the field be edited and the import panel open while a search runs", async () => {
		stubSearchAnswer("pending");
		renderAssets("/assets?q=PNL");
		expect(await screen.findByText("Searching…")).toBeInTheDocument();

		const field = screen.getByRole("textbox", { name: "Search assets" });
		fireEvent.change(field, { target: { value: "PNL-001" } });
		expect(field).toHaveValue("PNL-001");

		fireEvent.click(screen.getByRole("button", { name: "Import CSV" }));
		expect(getDialog()).toHaveAttribute("open");
	});
});
