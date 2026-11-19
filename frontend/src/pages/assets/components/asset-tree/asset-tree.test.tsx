import { fireEvent, render, screen, within } from "@testing-library/react";
import { Provider } from "react-redux";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { HTTPCode } from "~/libs/enums/enums.js";
import { config } from "~/libs/modules/config/config.js";
import { createStore } from "~/libs/modules/store/store.module.js";
import {
	type AssetListDto,
	AssetType,
	OperationalStatus,
} from "~/modules/assets/assets.js";

import { AssetTree } from "./asset-tree.js";

const TOP_LEVEL: AssetListDto = {
	items: [
		{
			assetId: "SS-001",
			assetName: "Aurora Substation",
			assetType: AssetType.SUBSTATION,
			childCount: 2,
			operationalStatus: OperationalStatus.IN_SERVICE,
		},
		{
			assetId: "SS-002",
			assetName: "Beacon Substation",
			assetType: AssetType.SUBSTATION,
			childCount: 0,
			operationalStatus: null,
		},
	],
	totalItems: 2,
};

const CHILDREN: AssetListDto = {
	items: [
		{
			assetId: "TX-001-1",
			assetName: "Aurora Transformer T1",
			assetType: AssetType.TRANSFORMER,
			childCount: 0,
			operationalStatus: OperationalStatus.MAINTENANCE,
		},
		{
			assetId: "SWB-001-1",
			assetName: "Aurora Switchboard A",
			assetType: AssetType.SWITCHBOARD,
			childCount: 1,
			operationalStatus: OperationalStatus.OUT_OF_SERVICE,
		},
	],
	totalItems: 2,
};

const NO_REQUESTS = 0;
const ONE_REQUEST = 1;
const REQUESTS_WITH_RETRY = 2;

const CHILDREN_URL = "http://localhost/api/assets?parentAssetId=SS-001";

const respond = (status: number, data: unknown): Promise<Response> =>
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

type ChildrenAnswer = "fail" | "ok" | "pending";

const stubFetch = (childrenAnswers: ChildrenAnswer[] = ["ok"]) => {
	const answers = [...childrenAnswers];
	const fetchMock = vi.fn((request: Request): Promise<Response> => {
		if (request.url !== CHILDREN_URL) {
			return respond(HTTPCode.OK, TOP_LEVEL);
		}

		const answer = answers.shift() ?? "ok";

		if (answer === "pending") {
			return new Promise<Response>(() => {
				// Never settles: the children stay loading.
			});
		}

		return answer === "fail"
			? respond(HTTPCode.INTERNAL_SERVER_ERROR, null)
			: respond(HTTPCode.OK, CHILDREN);
	});
	vi.stubGlobal("fetch", fetchMock);

	return fetchMock;
};

const countChildRequests = (fetchMock: ReturnType<typeof stubFetch>): number =>
	fetchMock.mock.calls.filter(([request]) => request.url === CHILDREN_URL)
		.length;

const renderTree = (): void => {
	render(
		<Provider store={createStore(config)}>
			<MemoryRouter>
				<AssetTree selectedAssetId={null} />
			</MemoryRouter>
		</Provider>,
	);
};

// A node's accessible name includes its descendants, so look it up by ID.
const getNode = (assetId: string): HTMLElement => {
	const node = screen
		.getAllByRole("treeitem")
		.find((element) => element.dataset["assetId"] === assetId);

	if (!node) {
		throw new Error(`No tree node ${assetId}`);
	}

	return node;
};

afterEach(() => {
	vi.unstubAllGlobals();
});

describe("AssetTree", () => {
	it("renders the top level from GET /api/assets", async () => {
		const fetchMock = stubFetch();
		renderTree();

		expect(await screen.findByText("SS-001")).toBeInTheDocument();
		expect(screen.getByRole("tree")).toBeInTheDocument();
		expect(screen.getByText("Beacon Substation")).toBeInTheDocument();
		expect(fetchMock.mock.calls.map(([request]) => request.url)).toEqual([
			"http://localhost/api/assets",
		]);
	});

	it("shows the loading state while the top level loads", () => {
		vi.stubGlobal(
			"fetch",
			vi.fn(
				() =>
					new Promise<Response>(() => {
						// Never settles.
					}),
			),
		);
		renderTree();

		expect(screen.getByText("Loading assets…")).toBeInTheDocument();
	});

	it("makes only nodes with children expandable", async () => {
		stubFetch();
		renderTree();
		await screen.findByText("SS-001");

		expect(getNode("SS-001")).toHaveAttribute("aria-expanded", "false");
		expect(getNode("SS-002")).not.toHaveAttribute("aria-expanded");
		expect(
			screen.queryByRole("button", { name: "Expand SS-002" }),
		).not.toBeInTheDocument();
	});

	it("loads children only when a node is expanded, and from the cache afterwards", async () => {
		const fetchMock = stubFetch();
		renderTree();
		await screen.findByText("SS-001");
		expect(countChildRequests(fetchMock)).toBe(NO_REQUESTS);

		fireEvent.click(screen.getByRole("button", { name: "Expand SS-001" }));

		expect(await screen.findByText("TX-001-1")).toBeInTheDocument();
		expect(getNode("SS-001")).toHaveAttribute("aria-expanded", "true");
		expect(countChildRequests(fetchMock)).toBe(ONE_REQUEST);

		fireEvent.click(screen.getByRole("button", { name: "Collapse SS-001" }));
		expect(screen.queryByText("TX-001-1")).not.toBeInTheDocument();

		fireEvent.click(screen.getByRole("button", { name: "Expand SS-001" }));
		expect(await screen.findByText("TX-001-1")).toBeInTheDocument();
		expect(countChildRequests(fetchMock)).toBe(ONE_REQUEST);
	});

	it("shows ID, name, type and a status label only when not in service", async () => {
		stubFetch();
		renderTree();
		await screen.findByText("SS-001");
		fireEvent.click(screen.getByRole("button", { name: "Expand SS-001" }));
		await screen.findByText("TX-001-1");

		const inService = getNode("SS-001");
		expect(
			within(inService).getByText("Aurora Substation"),
		).toBeInTheDocument();
		expect(within(inService).getByText("Substation")).toBeInTheDocument();
		expect(within(inService).queryByText("In service")).not.toBeInTheDocument();

		expect(
			within(getNode("SS-002")).getByText("No status recorded"),
		).toBeInTheDocument();
		expect(
			within(getNode("TX-001-1")).getByText("Maintenance"),
		).toBeInTheDocument();
		expect(
			within(getNode("TX-001-1")).getByText("Transformer"),
		).toBeInTheDocument();
		expect(
			within(getNode("SWB-001-1")).getByText("Out of service"),
		).toBeInTheDocument();
	});

	it("shows a loading row while a node's children load", async () => {
		stubFetch(["pending"]);
		renderTree();
		await screen.findByText("SS-001");

		fireEvent.click(screen.getByRole("button", { name: "Expand SS-001" }));

		expect(
			await within(getNode("SS-001")).findByText("Loading…"),
		).toBeInTheDocument();
	});

	it("shows a failed child load with a retry that refetches", async () => {
		const fetchMock = stubFetch(["fail", "ok"]);
		renderTree();
		await screen.findByText("SS-001");
		fireEvent.click(screen.getByRole("button", { name: "Expand SS-001" }));

		expect(
			await screen.findByText("Could not load children"),
		).toBeInTheDocument();

		fireEvent.click(screen.getByRole("button", { name: "Try again" }));

		expect(await screen.findByText("TX-001-1")).toBeInTheDocument();
		expect(countChildRequests(fetchMock)).toBe(REQUESTS_WITH_RETRY);
	});
});
