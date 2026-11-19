import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { DescendantCounts } from "./descendant-counts.js";

// The <dd> that belongs to a <dt> label.
const getValue = (label: string): HTMLElement => {
	const value = screen.getByText(label).nextElementSibling;

	if (!(value instanceof HTMLElement)) {
		throw new TypeError(`No value for ${label}`);
	}

	return value;
};

describe("DescendantCounts", () => {
	it("shows the four counts beneath a substation with their labels", () => {
		render(
			<DescendantCounts
				counts={{
					lvBoards: 1,
					switchboardPanels: 10,
					switchboards: 2,
					transformers: 2,
				}}
			/>,
		);

		expect(
			screen.getByRole("heading", { name: "Beneath this substation" }),
		).toBeInTheDocument();
		expect(getValue("Transformers")).toHaveTextContent("2");
		expect(getValue("LV boards")).toHaveTextContent("1");
		expect(getValue("Switchboards")).toHaveTextContent("2");
		expect(getValue("Switchboard panels")).toHaveTextContent("10");
	});

	it("shows zero counts as 0", () => {
		render(
			<DescendantCounts
				counts={{
					lvBoards: 0,
					switchboardPanels: 0,
					switchboards: 0,
					transformers: 0,
				}}
			/>,
		);

		expect(getValue("Switchboard panels")).toHaveTextContent("0");
	});
});
