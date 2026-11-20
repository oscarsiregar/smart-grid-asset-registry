import { render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { ImportStep } from "../../libs/enums/enums.js";
import { ImportSteps } from "./import-steps.js";

const getStep = (word: string): HTMLElement => {
	const step = screen
		.getAllByRole("listitem")
		.find((item) => item.textContent.includes(word));

	if (!step) {
		throw new Error(`No step ${word}`);
	}

	return step;
};

describe("ImportSteps", () => {
	it("lists the three steps in order with number and word", () => {
		render(<ImportSteps currentStep={ImportStep.CHOOSE_FILE} />);

		const list = screen.getByRole("list", { name: "Import steps" });
		expect(
			within(list)
				.getAllByRole("listitem")
				.map((item) => item.textContent),
		).toEqual(["1Choose file", "2Review", "3Done"]);
		expect(getStep("Choose file")).toHaveAttribute("aria-current", "step");
	});

	it("marks the current step and the earlier ones as done", () => {
		render(<ImportSteps currentStep={ImportStep.DONE} />);

		expect(getStep("Done")).toHaveAttribute("aria-current", "step");
		for (const word of ["Choose file", "Review"]) {
			const step = getStep(word);
			expect(step).not.toHaveAttribute("aria-current");
			expect(step.querySelector(".visually-hidden")).toHaveTextContent(
				"(done)",
			);
			expect(step.querySelector("svg")).toBeInTheDocument();
		}
	});

	it("marks nothing done before the current step", () => {
		render(<ImportSteps currentStep={ImportStep.REVIEW} />);

		expect(getStep("Review")).toHaveAttribute("aria-current", "step");
		expect(getStep("Done").querySelector(".visually-hidden")).toBeNull();
		expect(
			getStep("Choose file").querySelector(".visually-hidden"),
		).toHaveTextContent("(done)");
	});
});
