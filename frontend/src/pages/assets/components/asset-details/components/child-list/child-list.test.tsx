import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";

import {
	type AssetNodeDto,
	AssetType,
	OperationalStatus,
} from "~/modules/assets/assets.js";

import { ChildList } from "./child-list.js";

const PANELS: AssetNodeDto[] = [
	{
		assetId: "PNL-001-1-01",
		assetName: "Aurora Panel A1",
		assetType: AssetType.SWITCHBOARD_PANEL,
		childCount: 0,
		operationalStatus: OperationalStatus.MAINTENANCE,
	},
	{
		assetId: "PNL-001-1-02",
		assetName: "Aurora Panel A2",
		assetType: AssetType.SWITCHBOARD_PANEL,
		childCount: 0,
		operationalStatus: OperationalStatus.IN_SERVICE,
	},
];

describe("ChildList", () => {
	it("lists a switchboard's panels as links with ID, name and status, and their number", () => {
		render(
			<MemoryRouter>
				<ChildList childAssets={PANELS} label="Switchboard panels" />
			</MemoryRouter>,
		);

		expect(
			screen.getByRole("heading", { name: "Switchboard panels (2)" }),
		).toBeInTheDocument();
		const links = screen.getAllByRole("link");
		expect(links.map((link) => link.getAttribute("href"))).toEqual([
			"/assets/PNL-001-1-01",
			"/assets/PNL-001-1-02",
		]);
		const [firstLink] = links;
		expect(firstLink).toHaveTextContent("PNL-001-1-01");
		expect(firstLink).toHaveTextContent("Aurora Panel A1");
		expect(screen.getByText("Maintenance")).toBeInTheDocument();
		expect(screen.getByText("In service")).toBeInTheDocument();
	});
});
