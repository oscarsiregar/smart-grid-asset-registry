import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { afterEach } from "vitest";

/**
 * jsdom does not implement modal `<dialog>` behaviour. These methods only
 * toggle the `open` attribute so components using `showModal()` / `close()`
 * can render in tests; the real focus trap is checked in the browser.
 */
class DialogShim extends HTMLElement {
	public close(): void {
		this.removeAttribute("open");
	}

	public showModal(): void {
		this.setAttribute("open", "");
	}
}

const SHIMMED_DIALOG_METHODS = ["close", "showModal"] as const;

const installDialogShim = (): void => {
	for (const name of SHIMMED_DIALOG_METHODS) {
		const native = Object.getOwnPropertyDescriptor(
			HTMLDialogElement.prototype,
			name,
		);
		const shim = Object.getOwnPropertyDescriptor(DialogShim.prototype, name);

		if (shim && !native) {
			Object.defineProperty(HTMLDialogElement.prototype, name, shim);
		}
	}
};

installDialogShim();

/**
 * jsdom has no layout, so it does not implement `scrollIntoView`. A no-op
 * lets components call it; tests spy on it to check that they do.
 */
class ScrollShim extends HTMLElement {
	public override scrollIntoView(): void {
		// No layout to scroll in jsdom.
	}
}

const installScrollShim = (): void => {
	const native = Object.getOwnPropertyDescriptor(
		Element.prototype,
		"scrollIntoView",
	);
	const shim = Object.getOwnPropertyDescriptor(
		ScrollShim.prototype,
		"scrollIntoView",
	);

	if (shim && !native) {
		Object.defineProperty(Element.prototype, "scrollIntoView", shim);
	}
};

installScrollShim();

afterEach(() => {
	cleanup();
});
