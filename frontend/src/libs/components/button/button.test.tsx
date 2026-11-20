import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { IconName } from "~/libs/enums/enums.js";

import { Button } from "./button.js";

describe("Button", () => {
	it("shows its label by default", () => {
		render(<Button label="Close" type="button" />);

		const button = screen.getByRole("button", { name: "Close" });
		expect(button.querySelector(".visually-hidden")).toBeNull();
	});

	it("keeps a hidden label as the accessible name", () => {
		render(
			<Button
				iconName={IconName.CLOSE}
				isLabelHidden
				label="Close"
				type="button"
			/>,
		);

		const button = screen.getByRole("button", { name: "Close" });
		expect(button.querySelector(".visually-hidden")).toHaveTextContent("Close");
		expect(button.querySelector("svg")).toBeInTheDocument();
	});
});
