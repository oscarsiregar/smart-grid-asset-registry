import { render, screen, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";

import {
	type AssetNodeDto,
	AssetType,
	OperationalStatus,
} from "~/modules/assets/assets.js";

import { ChildGroups } from "./child-groups.js";

// Transformers, LV boards, switchboards.
const GROUP_COUNT = 3;

// Ordered as the API sends them: by type, then by ID.
const CHILD_ASSETS: AssetNodeDto[] = [
	{
		assetId: "TX-001-1",
		assetName: "Aurora Transformer T1",
		assetType: AssetType.TRANSFORMER,
		childCount: 0,
		operationalStatus: OperationalStatus.IN_SERVICE,
	},
	{
		assetId: "TX-001-2",
		assetName: "Aurora Transformer T2",
		assetType: AssetType.TRANSFORMER,
		childCount: 0,
		operationalStatus: OperationalStatus.MAINTENANCE,
	},
	{
		assetId: "SWB-001-1",
		assetName: "Aurora Switchboard A",
		assetType: AssetType.SWITCHBOARD,
		childCount: 4,
		operationalStatus: null,
	},
];

const renderGroups = (childAssets: AssetNodeDto[]): void => {
	render(
		<MemoryRouter>
			<ChildGroups childAssets={childAssets} />
		</MemoryRouter>,
	);
};

const getGroup = (heading: string): HTMLElement => {
	const group = screen.getByRole("heading", { name: heading }).parentElement;

	if (!group) {
		throw new Error(`No group ${heading}`);
	}

	return group;
};

describe("ChildGroups", () => {
	it("shows the three groups in a fixed order, each with its number of children", () => {
		renderGroups(CHILD_ASSETS);

		expect(
			screen.getAllByRole("heading").map((heading) => heading.textContent),
		).toEqual(["Transformers (2)", "LV boards (0)", "Switchboards (1)"]);
	});

	it("lists each child as a link with its ID, name and status in words", () => {
		renderGroups(CHILD_ASSETS);

		const transformers = getGroup("Transformers (2)");
		const links = within(transformers).getAllByRole("link");
		expect(links.map((link) => link.getAttribute("href"))).toEqual([
			"/assets/TX-001-1",
			"/assets/TX-001-2",
		]);
		const [firstLink] = links;
		expect(firstLink).toHaveTextContent("Aurora Transformer T1");
		expect(within(transformers).getByText("In service")).toBeInTheDocument();
		expect(within(transformers).getByText("Maintenance")).toBeInTheDocument();
		expect(
			within(getGroup("Switchboards (1)")).getByText("No status recorded"),
		).toBeInTheDocument();
	});

	it("states that an empty group has none", () => {
		renderGroups(CHILD_ASSETS);

		expect(
			within(getGroup("LV boards (0)")).getByText("None"),
		).toBeInTheDocument();
	});

	it("shows all three groups as empty for a substation without children", () => {
		renderGroups([]);

		expect(screen.getAllByText("None")).toHaveLength(GROUP_COUNT);
	});
});
