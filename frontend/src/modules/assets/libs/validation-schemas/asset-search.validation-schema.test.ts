import { describe, expect, it } from "vitest";

import {
	assetSearchValidationSchema as assetSearch,
	AssetType,
} from "../../assets.js";

const WHITESPACE_ONLY = " \t ";
// Longer than the backend accepts; the form leaves the length to it.
const LONG_TEXT_LENGTH = 500;

const getFirstMessage = (value: unknown): string | undefined => {
	const result = assetSearch.safeParse(value);
	const [firstIssue] = result.error?.issues ?? [];

	return firstIssue?.message;
};

describe("assetSearch", () => {
	it("trims the text", () => {
		const result = assetSearch.safeParse({ assetType: null, q: "  PNL-001  " });

		expect(result.success).toBe(true);
		expect(result.data?.q).toBe("PNL-001");
	});

	it.each(["", WHITESPACE_ONLY])("refuses a blank text (%j)", (q) => {
		expect(getFirstMessage({ assetType: null, q })).toBe(
			"Enter an asset ID or name to search.",
		);
	});

	it("leaves the length limit to the backend (principle II)", () => {
		expect(
			assetSearch.safeParse({
				assetType: null,
				q: "a".repeat(LONG_TEXT_LENGTH),
			}).success,
		).toBe(true);
	});

	it.each([...Object.values(AssetType), null])(
		"accepts the asset type %s",
		(assetType) => {
			expect(assetSearch.safeParse({ assetType, q: "PNL" }).success).toBe(true);
		},
	);

	it("leaves everything else to the backend", () => {
		expect(
			assetSearch.safeParse({ assetType: null, q: String.raw`%_\ / ?` })
				.success,
		).toBe(true);
	});
});
