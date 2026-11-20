import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { TreeSkeleton } from "./tree-skeleton.js";

const ROWS = 3;

describe("TreeSkeleton", () => {
	it("announces its text and shows placeholder rows hidden from assistive technology", () => {
		const { container } = render(<TreeSkeleton label="Loading…" rows={ROWS} />);

		const status = screen.getByRole("status");
		expect(status).toHaveTextContent("Loading…");
		expect(screen.getByText("Loading…")).toHaveClass("visually-hidden");

		const rows = container.querySelectorAll("[data-skeleton-row]");
		expect(rows).toHaveLength(ROWS);

		for (const row of rows) {
			expect(row).toHaveAttribute("aria-hidden", "true");
			expect(row.querySelector("[data-shape='dot']")).toBeInTheDocument();
			expect(row.querySelector("[data-shape='text']")).toBeInTheDocument();
		}
	});
});
