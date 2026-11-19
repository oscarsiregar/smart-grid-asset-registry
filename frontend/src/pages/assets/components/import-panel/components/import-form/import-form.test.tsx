import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ImportForm } from "./import-form.js";

const CHANGE_COUNT = 2;

const createCsvFile = (): File =>
	new File(["asset_id\n"], "grid_assets.csv", { type: "text/csv" });

const chooseFile = (file: File): void => {
	fireEvent.change(screen.getByLabelText("CSV file"), {
		target: { files: [file] },
	});
};

describe("ImportForm", () => {
	it("starts in Partial mode with Check file disabled until a file is chosen", () => {
		render(
			<ImportForm isLoading={false} onChange={vi.fn()} onSubmit={vi.fn()} />,
		);

		expect(screen.getByRole("radio", { name: /Partial/ })).toBeChecked();
		expect(screen.getByRole("radio", { name: /Strict/ })).not.toBeChecked();
		expect(screen.getByRole("button", { name: "Check file" })).toBeDisabled();
	});

	it("explains each mode in one line", () => {
		render(
			<ImportForm isLoading={false} onChange={vi.fn()} onSubmit={vi.fn()} />,
		);

		expect(
			screen.getByRole("radio", { name: /Strict/ }),
		).toHaveAccessibleDescription("Save nothing if any row is rejected.");
	});

	it("submits the chosen file and mode", async () => {
		const handleSubmit = vi.fn();
		render(
			<ImportForm
				isLoading={false}
				onChange={vi.fn()}
				onSubmit={handleSubmit}
			/>,
		);
		const file = createCsvFile();

		chooseFile(file);
		fireEvent.click(screen.getByRole("radio", { name: /Strict/ }));
		fireEvent.click(screen.getByRole("button", { name: "Check file" }));

		await waitFor(() => {
			expect(handleSubmit).toHaveBeenCalledOnce();
		});
		expect(handleSubmit).toHaveBeenCalledWith(
			{ file, mode: "STRICT" },
			expect.anything(),
		);
	});

	it("reports every change of file or mode so the shown result can be cleared", () => {
		const handleChange = vi.fn();
		render(
			<ImportForm
				isLoading={false}
				onChange={handleChange}
				onSubmit={vi.fn()}
			/>,
		);

		chooseFile(createCsvFile());
		fireEvent.click(screen.getByRole("radio", { name: /Strict/ }));

		expect(handleChange).toHaveBeenCalledTimes(CHANGE_COUNT);
	});

	it("shows that the file is being checked", () => {
		render(<ImportForm isLoading onChange={vi.fn()} onSubmit={vi.fn()} />);

		expect(screen.getByText("Checking file…")).toHaveAttribute(
			"role",
			"status",
		);
	});
});
