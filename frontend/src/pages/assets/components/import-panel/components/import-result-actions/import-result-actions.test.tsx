import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ImportResultActions } from "./import-result-actions.js";

const NEW_ASSETS = 205;

describe("ImportResultActions", () => {
	it("offers to choose another file", () => {
		const handleReset = vi.fn();
		render(
			<ImportResultActions
				confirmCount={NEW_ASSETS}
				onReset={handleReset}
				resetLabel="Choose another file"
			/>,
		);

		fireEvent.click(
			screen.getByRole("button", { name: "Choose another file" }),
		);

		expect(handleReset).toHaveBeenCalledOnce();
	});

	it("offers Save with the number of new assets when confirming is possible", () => {
		const handleConfirm = vi.fn();
		render(
			<ImportResultActions
				confirmCount={NEW_ASSETS}
				onConfirm={handleConfirm}
				onReset={vi.fn()}
				resetLabel="Choose another file"
			/>,
		);

		fireEvent.click(screen.getByRole("button", { name: "Save 205 assets" }));

		expect(handleConfirm).toHaveBeenCalledOnce();
	});

	it("shows that the import is being saved and blocks both actions", () => {
		render(
			<ImportResultActions
				confirmCount={NEW_ASSETS}
				isSaving
				onConfirm={vi.fn()}
				onReset={vi.fn()}
				resetLabel="Choose another file"
			/>,
		);

		expect(screen.getByText("Saving…")).toHaveAttribute("role", "status");
		expect(
			screen.getByRole("button", { name: /Save 205 assets/ }),
		).toBeDisabled();
		expect(
			screen.getByRole("button", { name: "Choose another file" }),
		).toBeDisabled();
	});

	it("offers no Save without onConfirm", () => {
		render(
			<ImportResultActions
				confirmCount={NEW_ASSETS}
				onReset={vi.fn()}
				resetLabel="Import another file"
			/>,
		);

		expect(
			screen.queryByRole("button", { name: /Save/ }),
		).not.toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: "Import another file" }),
		).toBeEnabled();
	});
});
