import { fireEvent, render, screen, within } from "@testing-library/react";
import { Provider } from "react-redux";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { HTTPCode } from "~/libs/enums/enums.js";
import { config } from "~/libs/modules/config/config.js";
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
});
