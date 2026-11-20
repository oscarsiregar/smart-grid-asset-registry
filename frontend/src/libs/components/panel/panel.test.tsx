import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { Panel } from "./panel.js";

const PANEL_TITLE = "Import assets from CSV";
const PANEL_DESCRIPTION =
	"Validate a file, review what would be rejected, then save the valid rows.";

type HarnessProperties = {
	isOpen: boolean;
	onClose: () => void;
};

const Harness: React.FC<HarnessProperties> = ({
	isOpen,
	onClose,
}: HarnessProperties) => (
	<>
		<button type="button">Opener</button>
		<Panel isOpen={isOpen} onClose={onClose} title={PANEL_TITLE}>
			<p>Panel body</p>
		</Panel>
	</>
);

// A closed <dialog> is hidden, so it is looked up with { hidden: true }.
const getDialog = (): HTMLElement =>
	screen.getByRole("dialog", { hidden: true });

describe("Panel", () => {
	it("opens and closes with isOpen, mounting children only while open", () => {
		const { rerender } = render(<Harness isOpen={false} onClose={vi.fn()} />);

		expect(getDialog()).not.toHaveAttribute("open");
		expect(screen.queryByText("Panel body")).not.toBeInTheDocument();

		rerender(<Harness isOpen onClose={vi.fn()} />);

		expect(getDialog()).toHaveAttribute("open");
		expect(screen.getByText("Panel body")).toBeInTheDocument();

		rerender(<Harness isOpen={false} onClose={vi.fn()} />);

		expect(getDialog()).not.toHaveAttribute("open");
		expect(screen.queryByText("Panel body")).not.toBeInTheDocument();
	});

	it("is labelled by its title", () => {
		render(<Harness isOpen onClose={vi.fn()} />);

		expect(getDialog()).toHaveAccessibleName(PANEL_TITLE);
	});

	it("is described by its subtitle when it has one", () => {
		render(
			<Panel
				description={PANEL_DESCRIPTION}
				isOpen
				onClose={vi.fn()}
				title={PANEL_TITLE}
			>
				<p>Panel body</p>
			</Panel>,
		);

		expect(screen.getByText(PANEL_DESCRIPTION)).toBeInTheDocument();
		expect(getDialog()).toHaveAccessibleDescription(PANEL_DESCRIPTION);
	});

	it("shows only an icon on the Close button, still named Close", () => {
		render(<Harness isOpen onClose={vi.fn()} />);

		const close = screen.getByRole("button", { name: "Close" });
		expect(close.querySelector(".visually-hidden")).toHaveTextContent("Close");
		expect(close.querySelector("svg")).toBeInTheDocument();
	});

	it("calls onClose on Escape (cancel) and keeps itself open until told otherwise", () => {
		const handleClose = vi.fn();
		render(<Harness isOpen onClose={handleClose} />);

		// Browsers fire a cancelable "cancel" event on Escape.
		const cancelEvent = new Event("cancel", { cancelable: true });
		fireEvent(getDialog(), cancelEvent);

		expect(handleClose).toHaveBeenCalledOnce();
		expect(cancelEvent.defaultPrevented).toBe(true);
		expect(getDialog()).toHaveAttribute("open");
	});

	it("calls onClose from the Close button", () => {
		const handleClose = vi.fn();
		render(<Harness isOpen onClose={handleClose} />);

		fireEvent.click(screen.getByRole("button", { name: "Close" }));

		expect(handleClose).toHaveBeenCalledOnce();
	});

	it("returns focus to the element that opened it", () => {
		const { rerender } = render(<Harness isOpen={false} onClose={vi.fn()} />);
		const opener = screen.getByRole("button", { name: "Opener" });
		opener.focus();

		rerender(<Harness isOpen onClose={vi.fn()} />);
		screen.getByRole("button", { name: "Close" }).focus();
		rerender(<Harness isOpen={false} onClose={vi.fn()} />);

		expect(opener).toHaveFocus();
	});
});
