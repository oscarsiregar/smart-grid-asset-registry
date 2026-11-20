import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { DetailsSkeleton } from "./details-skeleton.js";

// The details card's placeholder attributes: 2 rows of 5 (spec 005 contract).
const ATTRIBUTE_PLACEHOLDERS = 10;
const CHILD_ROW_PLACEHOLDERS = 3;

describe("DetailsSkeleton", () => {
	it("announces Loading asset… and imitates the intro and both cards", () => {
		const { container } = render(<DetailsSkeleton />);

		expect(screen.getByRole("status")).toHaveTextContent("Loading asset…");
		expect(
			container.querySelector(
				":scope [data-skeleton-part='intro'] [data-shape='title']",
			),
		).toBeInTheDocument();
		expect(
			container.querySelectorAll("[data-skeleton-part='attribute']"),
		).toHaveLength(ATTRIBUTE_PLACEHOLDERS);
		expect(
			container.querySelectorAll("[data-skeleton-part='child-row']"),
		).toHaveLength(CHILD_ROW_PLACEHOLDERS);
		for (const part of container.querySelectorAll("[data-skeleton-part]")) {
			expect(part.closest("[aria-hidden='true']")).not.toBeNull();
		}
	});
});
