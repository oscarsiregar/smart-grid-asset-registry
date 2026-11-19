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

afterEach(() => {
	cleanup();
});
