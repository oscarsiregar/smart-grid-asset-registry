import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { ImportOutcomeMessage, OutcomeTone } from "../../libs/enums/enums.js";
import { OutcomeBanner } from "./outcome-banner.js";

describe("OutcomeBanner", () => {
	it.each([
		[ImportOutcomeMessage.VALIDATED_ONLY, OutcomeTone.INFO],
		[ImportOutcomeMessage.CHECK_FAILED, OutcomeTone.DANGER],
		["Not committed — strict mode, 17 rows rejected", OutcomeTone.WARNING],
	])("states %s as text in a status region", (message, tone) => {
		render(<OutcomeBanner message={message} tone={tone} />);

		expect(screen.getByRole("status")).toHaveTextContent(message);
	});

	it("pairs the tone with an icon that is decorative, so the text carries the meaning", () => {
		const { container } = render(
			<OutcomeBanner
				message={ImportOutcomeMessage.CHECK_FAILED}
				tone={OutcomeTone.DANGER}
			/>,
		);

		const icon = container.querySelector("svg");
		expect(icon).toHaveAttribute("aria-hidden", "true");
		expect(screen.getByRole("status")).toHaveTextContent(
			"Not checked — the file could not be checked. Nothing was saved.",
		);
	});
});
