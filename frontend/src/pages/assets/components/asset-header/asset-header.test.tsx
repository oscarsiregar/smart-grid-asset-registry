import { render, screen } from "@testing-library/react";
import { Provider } from "react-redux";
import { afterEach, describe, expect, it, vi } from "vitest";

import { HTTPCode } from "~/libs/enums/enums.js";
import { config } from "~/libs/modules/config/config.js";
import { createStore } from "~/libs/modules/store/store.module.js";
import {
	type AssetDetailDto,
	AssetType,
	OperationalStatus,
} from "~/modules/assets/assets.js";

import { AssetHeader } from "./asset-header.js";

const DESCRIPTION = "Substations and the equipment installed in them.";

const TRANSFORMER: AssetDetailDto = {
	asset: {
		assetId: "TX-001-1",
		assetName: "Aurora Transformer T1",
		assetType: AssetType.TRANSFORMER,
		commissionedDate: null,
		manufacturer: null,
		model: null,
		operationalStatus: OperationalStatus.IN_SERVICE,
		parentAssetId: "SS-001",
		ratingKva: null,
		serialNumber: null,
		voltageKv: null,
	},
	children: [],
	descendantCounts: null,
	path: [
		{
			assetId: "SS-001",
			assetName: "Aurora Substation (North)",
			assetType: AssetType.SUBSTATION,
			childCount: 5,
			operationalStatus: OperationalStatus.IN_SERVICE,
		},
		{
			assetId: "TX-001-1",
			assetName: "Aurora Transformer T1",
			assetType: AssetType.TRANSFORMER,
			childCount: 0,
			operationalStatus: OperationalStatus.IN_SERVICE,
		},
	],
};

const stubDetail = (status: number, data: unknown): void => {
	vi.stubGlobal(
		"fetch",
		vi.fn(() =>
			Promise.resolve(
				Response.json(
					{
						code: status,
						data,
						message: status === HTTPCode.OK ? "ok" : "Internal Server Error",
						status: status === HTTPCode.OK ? "success" : "error",
					},
					{ status },
				),
			),
		),
	);
};

const renderHeader = (selectedAssetId: null | string): void => {
	render(
		<Provider store={createStore(config)}>
			<AssetHeader selectedAssetId={selectedAssetId} />
		</Provider>,
	);
};

afterEach(() => {
	vi.unstubAllGlobals();
});

describe("AssetHeader", () => {
	it("reads Asset explorer with the description when nothing is selected", () => {
		const fetchMock = vi.fn();
		vi.stubGlobal("fetch", fetchMock);
		renderHeader(null);

		expect(
			screen.getByRole("heading", { level: 1, name: "Asset explorer" }),
		).toBeInTheDocument();
		expect(screen.getByText(DESCRIPTION)).toBeInTheDocument();
		expect(fetchMock).not.toHaveBeenCalled();
	});

	it("shows the selected asset's path names as plain text", async () => {
		stubDetail(HTTPCode.OK, TRANSFORMER);
		renderHeader("TX-001-1");

		expect(
			await screen.findByText(
				"Aurora Substation (North) / Aurora Transformer T1",
			),
		).toBeInTheDocument();
		expect(screen.queryByRole("link")).not.toBeInTheDocument();
		expect(screen.queryByText(DESCRIPTION)).not.toBeInTheDocument();
	});

	it("keeps the description while the details load or fail", async () => {
		stubDetail(HTTPCode.INTERNAL_SERVER_ERROR, null);
		renderHeader("TX-001-1");

		expect(screen.getByText(DESCRIPTION)).toBeInTheDocument();
		await vi.waitFor(() => {
			expect(screen.getByText(DESCRIPTION)).toBeInTheDocument();
		});
	});
});
