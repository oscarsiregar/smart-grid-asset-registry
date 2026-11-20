import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { SkeletonShape, SkeletonWidth } from "./libs/enums/enums.js";
import { Skeleton } from "./skeleton.js";

const renderBlock = (element: React.ReactElement): HTMLElement => {
	const { container } = render(element);
	const block = container.firstElementChild;

	if (!(block instanceof HTMLElement)) {
		throw new TypeError("Expected the skeleton block");
	}

	return block;
};

describe("Skeleton", () => {
	it("is one decorative block hidden from assistive technology", () => {
		const block = renderBlock(<Skeleton />);

		expect(block).toHaveAttribute("aria-hidden", "true");
		expect(block).toBeEmptyDOMElement();
	});

	it("defaults to a full-width line of text", () => {
		const block = renderBlock(<Skeleton />);

		expect(block).toHaveAttribute("data-shape", SkeletonShape.TEXT);
		expect(block).toHaveAttribute("data-width", SkeletonWidth.FULL);
	});

	it("takes the given shape and width", () => {
		const block = renderBlock(
			<Skeleton shape={SkeletonShape.TITLE} width={SkeletonWidth.HALF} />,
		);

		expect(block).toHaveAttribute("data-shape", SkeletonShape.TITLE);
		expect(block).toHaveAttribute("data-width", SkeletonWidth.HALF);
	});

	it("merges a class name", () => {
		const block = renderBlock(<Skeleton className="extra" />);

		expect(block).toHaveClass("extra");
	});
});
