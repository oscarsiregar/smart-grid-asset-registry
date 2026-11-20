import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { APP_NAME } from "~/libs/constants/constants.js";

import { Logo } from "./logo.js";

describe("Logo", () => {
	it("shows the bolt mark in a decorative tile beside the application name", () => {
		const { container } = render(<Logo />);

		expect(screen.getByText(APP_NAME)).toBeInTheDocument();
		const tile = container.querySelector("[data-logo-mark]");
		expect(tile).toHaveAttribute("aria-hidden", "true");
		expect(tile?.querySelector("svg")).toBeInTheDocument();
		expect(container.querySelector("img")).toBeNull();
	});

	it("names the mark when it is shown alone", () => {
		render(<Logo isIconOnly />);

		expect(
			screen.getByRole("img", { name: `${APP_NAME} logo` }),
		).toBeInTheDocument();
		expect(screen.queryByText(APP_NAME)).not.toBeInTheDocument();
	});
});
