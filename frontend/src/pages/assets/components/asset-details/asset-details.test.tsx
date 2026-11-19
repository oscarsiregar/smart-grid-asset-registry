import { fireEvent, render, screen, within } from "@testing-library/react";
import { Provider } from "react-redux";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { HTTPCode } from "~/libs/enums/enums.js";
import { config } from "~/libs/modules/config/config.js";
import { bindShowNotification } from "~/libs/modules/notification/notification.js";
import { createStore } from "~/libs/modules/store/store.module.js";
import {
	type AssetDetailDto,
	AssetType,
	OperationalStatus,
} from "~/modules/assets/assets.js";

import { AssetDetails } from "./asset-details.js";

const PANEL: AssetDetailDto = {
	asset: {
		assetId: "PNL-001-1-01",
		assetName: "Aurora Panel A1",
		assetType: AssetType.SWITCHBOARD_PANEL,
		commissionedDate: "2010-09-11",
		manufacturer: null,
		model: null,
		operationalStatus: null,
		parentAssetId: "SWB-001-1",
		ratingKva: 1000,
		serialNumber: "PN0010101",
		voltageKv: 22,
	},
	children: [],
	descendantCounts: null,
	path: [
		{
			assetId: "SS-001",
			assetName: "Aurora Substation",
			assetType: AssetType.SUBSTATION,
			childCount: 5,
			operationalStatus: OperationalStatus.IN_SERVICE,
		},
		{
			assetId: "SWB-001-1",
			assetName: "Aurora Switchboard A",
			assetType: AssetType.SWITCHBOARD,
			childCount: 1,
			operationalStatus: OperationalStatus.OUT_OF_SERVICE,
		},
		{
			assetId: "PNL-001-1-01",
			assetName: "Aurora Panel A1",
			assetType: AssetType.SWITCHBOARD_PANEL,
			childCount: 0,
			operationalStatus: null,
		},
	],
};

// The failed request plus the retry.
const REQUESTS_WITH_RETRY = 2;

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

const renderDetails = (selectedAssetId: null | string): void => {
	render(
		<Provider store={createStore(config)}>
			<MemoryRouter>
				<AssetDetails selectedAssetId={selectedAssetId} />
			</MemoryRouter>
		</Provider>,
	);
};

// The <dd> that belongs to a <dt> label.
const getValue = (label: string): HTMLElement => {
	const value = screen.getByText(label).nextElementSibling;

	if (!(value instanceof HTMLElement)) {
		throw new TypeError(`No value for ${label}`);
	}

	return value;
};

afterEach(() => {
	vi.unstubAllGlobals();
});

describe("AssetDetails", () => {
	it("invites a selection when no asset is selected", () => {
		const fetchMock = vi.fn();
		vi.stubGlobal("fetch", fetchMock);

		renderDetails(null);

		expect(
			screen.getByText("Select an asset in the tree to see its details."),
		).toBeInTheDocument();
		expect(fetchMock).not.toHaveBeenCalled();
	});

	it("says the asset is loading", () => {
		vi.stubGlobal(
			"fetch",
			vi.fn(
				() =>
					new Promise<Response>(() => {
						// Never settles.
					}),
			),
		);

		renderDetails("PNL-001-1-01");

		expect(screen.getByRole("status")).toHaveTextContent("Loading asset…");
	});

	it("shows the path, the heading and every attribute of a panel", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn(() => respondWith(HTTPCode.OK, PANEL)),
		);

		renderDetails("PNL-001-1-01");

		expect(
			await screen.findByRole("heading", { name: "Aurora Panel A1" }),
		).toBeInTheDocument();

		const path = screen.getByRole("navigation", { name: "Path" });
		expect(within(path).getByRole("link", { name: "SS-001" })).toHaveAttribute(
			"href",
			"/assets/SS-001",
		);
		expect(
			within(path).getByRole("link", { name: "SWB-001-1" }),
		).toHaveAttribute("href", "/assets/SWB-001-1");
		expect(
			within(path).queryByRole("link", { name: "PNL-001-1-01" }),
		).not.toBeInTheDocument();
		expect(within(path).getByText("PNL-001-1-01")).toHaveAttribute(
			"aria-current",
			"page",
		);

		expect(getValue("Asset ID")).toHaveTextContent("PNL-001-1-01");
		expect(getValue("Type")).toHaveTextContent("Switchboard panel");
		expect(getValue("Name")).toHaveTextContent("Aurora Panel A1");
		expect(getValue("Operational status")).toHaveTextContent(
			"No status recorded",
		);
		expect(getValue("Voltage")).toHaveTextContent("22 kV");
		expect(getValue("Rating")).toHaveTextContent("1000 kVA");
		expect(getValue("Manufacturer")).toHaveTextContent("Not recorded");
		expect(getValue("Model")).toHaveTextContent("Not recorded");
		expect(getValue("Serial number")).toHaveTextContent("PN0010101");
		expect(getValue("Commissioned date")).toHaveTextContent("2010-09-11");

		const parent = within(getValue("Parent asset")).getByRole("link");
		expect(parent).toHaveAttribute("href", "/assets/SWB-001-1");
		expect(parent).toHaveTextContent("SWB-001-1");
		expect(parent).toHaveTextContent("Aurora Switchboard A");
	});

	it("shows blank numbers and dates as not recorded, and no parent for a substation", async () => {
		const [substationStep] = PANEL.path;

		if (!substationStep) {
			throw new Error("The fixture has no substation");
		}

		const substation: AssetDetailDto = {
			asset: {
				...PANEL.asset,
				assetId: "SS-001",
				assetName: "Aurora Substation",
				assetType: AssetType.SUBSTATION,
				commissionedDate: null,
				operationalStatus: OperationalStatus.IN_SERVICE,
				parentAssetId: null,
				ratingKva: null,
				voltageKv: null,
			},
			children: [],
			descendantCounts: {
				lvBoards: 0,
				switchboardPanels: 0,
				switchboards: 0,
				transformers: 0,
			},
			path: [substationStep],
		};
		vi.stubGlobal(
			"fetch",
			vi.fn(() => respondWith(HTTPCode.OK, substation)),
		);

		renderDetails("SS-001");

		await screen.findByRole("heading", { name: "Aurora Substation" });
		expect(getValue("Operational status")).toHaveTextContent("In service");
		expect(getValue("Voltage")).toHaveTextContent("Not recorded");
		expect(getValue("Rating")).toHaveTextContent("Not recorded");
		expect(getValue("Commissioned date")).toHaveTextContent("Not recorded");
		expect(screen.queryByText("Parent asset")).not.toBeInTheDocument();
	});

	it("shows a failure with a retry that refetches", async () => {
		const fetchMock = vi
			.fn()
			.mockImplementationOnce(() =>
				respondWith(HTTPCode.INTERNAL_SERVER_ERROR, null),
			)
			.mockImplementation(() => respondWith(HTTPCode.OK, PANEL));
		vi.stubGlobal("fetch", fetchMock);

		renderDetails("PNL-001-1-01");

		expect(
			await screen.findByText("The asset could not be loaded."),
		).toBeInTheDocument();

		fireEvent.click(screen.getByRole("button", { name: "Try again" }));

		expect(
			await screen.findByRole("heading", { name: "Aurora Panel A1" }),
		).toBeInTheDocument();
		expect(fetchMock).toHaveBeenCalledTimes(REQUESTS_WITH_RETRY);
	});
	it("shows a substation's child groups and the counts beneath it", async () => {
		const [substationStep, switchboardStep] = PANEL.path;

		if (!substationStep || !switchboardStep) {
			throw new Error("The fixture has no substation or switchboard");
		}

		const substation: AssetDetailDto = {
			asset: {
				...PANEL.asset,
				assetId: "SS-001",
				assetName: "Aurora Substation",
				assetType: AssetType.SUBSTATION,
				parentAssetId: null,
			},
			children: [switchboardStep],
			descendantCounts: {
				lvBoards: 1,
				switchboardPanels: 10,
				switchboards: 2,
				transformers: 2,
			},
			path: [substationStep],
		};
		vi.stubGlobal(
			"fetch",
			vi.fn(() => respondWith(HTTPCode.OK, substation)),
		);

		renderDetails("SS-001");

		expect(
			await screen.findByRole("heading", { name: "Switchboards (1)" }),
		).toBeInTheDocument();
		expect(
			screen.getByRole("heading", { name: "Beneath this substation" }),
		).toBeInTheDocument();
		expect(
			screen.queryByRole("heading", { name: /^Switchboard panels/ }),
		).not.toBeInTheDocument();
	});

	it("lists a switchboard's panels", async () => {
		const [substationStep, switchboardStep, panelStep] = PANEL.path;

		if (!substationStep || !switchboardStep || !panelStep) {
			throw new Error("The fixture path is incomplete");
		}

		const switchboard: AssetDetailDto = {
			asset: {
				...PANEL.asset,
				assetId: "SWB-001-1",
				assetName: "Aurora Switchboard A",
				assetType: AssetType.SWITCHBOARD,
				parentAssetId: "SS-001",
			},
			children: [panelStep],
			descendantCounts: null,
			path: [substationStep, switchboardStep],
		};
		vi.stubGlobal(
			"fetch",
			vi.fn(() => respondWith(HTTPCode.OK, switchboard)),
		);

		renderDetails("SWB-001-1");

		expect(
			await screen.findByRole("heading", { name: "Switchboard panels (1)" }),
		).toBeInTheDocument();
		expect(
			screen.queryByRole("heading", { name: "Beneath this substation" }),
		).not.toBeInTheDocument();
	});

	it("shows neither groups nor a child list for a panel", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn(() => respondWith(HTTPCode.OK, PANEL)),
		);

		renderDetails("PNL-001-1-01");

		await screen.findByRole("heading", { name: "Aurora Panel A1" });
		expect(
			screen.queryByRole("heading", {
				name: /^(Transformers|Switchboard panels)/,
			}),
		).not.toBeInTheDocument();
		expect(
			screen.queryByRole("heading", { name: "Beneath this substation" }),
		).not.toBeInTheDocument();
	});
	it("shows ASSET_NOT_FOUND as a not-found state, not an error, without a toast", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn(() =>
				Promise.resolve(
					Response.json(
						{
							code: HTTPCode.NOT_FOUND,
							data: { errorCode: "ASSET_NOT_FOUND" },
							message: "No asset with ID NOPE exists.",
							status: "error",
						},
						{ status: HTTPCode.NOT_FOUND },
					),
				),
			),
		);
		const showNotificationMock = vi.fn();
		const unbind = bindShowNotification(showNotificationMock);

		renderDetails("NOPE");

		expect(
			await screen.findByText("No asset with ID NOPE exists."),
		).toBeInTheDocument();
		expect(
			screen.getByRole("link", { name: "Back to all assets" }),
		).toHaveAttribute("href", "/assets");
		expect(
			screen.queryByText("The asset could not be loaded."),
		).not.toBeInTheDocument();
		expect(
			screen.queryByRole("button", { name: "Try again" }),
		).not.toBeInTheDocument();
		expect(showNotificationMock).not.toHaveBeenCalled();
		unbind();
	});

	it("keeps a 404 without ASSET_NOT_FOUND as an error", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn(() =>
				Promise.resolve(
					Response.json(
						{
							code: HTTPCode.NOT_FOUND,
							data: null,
							message: "Not Found",
							status: "error",
						},
						{ status: HTTPCode.NOT_FOUND },
					),
				),
			),
		);

		renderDetails("NOPE");

		expect(
			await screen.findByText("The asset could not be loaded."),
		).toBeInTheDocument();
		expect(
			screen.queryByText("No asset with ID NOPE exists."),
		).not.toBeInTheDocument();
	});
});
