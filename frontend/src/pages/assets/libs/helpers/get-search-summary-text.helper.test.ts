import { describe, expect, it } from "vitest";

import {
	type AssetSearchRequestDto,
	AssetType,
} from "~/modules/assets/assets.js";

import { getSearchSummaryText } from "./get-search-summary-text.helper.js";

const TEXT_ONLY: AssetSearchRequestDto = { assetType: null, q: "pnl" };
const WITH_TYPE: AssetSearchRequestDto = {
	assetType: AssetType.SWITCHBOARD_PANEL,
	q: "pnl",
};

const NONE = 0;
const ONE = 1;
const SEVERAL = 12;
const RECEIVED = 50;
const MANY = 1234;

type Case = {
	expected: string;
	receivedCount: number;
	search: AssetSearchRequestDto;
	totalItems: number;
};

const CASES: Case[] = [
	{
		expected: "No asset matches “pnl”",
		receivedCount: NONE,
		search: TEXT_ONLY,
		totalItems: NONE,
	},
	{
		expected: "No asset matches “pnl” of type Switchboard panel",
		receivedCount: NONE,
		search: WITH_TYPE,
		totalItems: NONE,
	},
	{
		expected: "1 asset matches “pnl”",
		receivedCount: ONE,
		search: TEXT_ONLY,
		totalItems: ONE,
	},
	{
		expected: "12 assets match “pnl”",
		receivedCount: SEVERAL,
		search: TEXT_ONLY,
		totalItems: SEVERAL,
	},
	{
		expected: "12 assets match “pnl” of type Switchboard panel",
		receivedCount: SEVERAL,
		search: WITH_TYPE,
		totalItems: SEVERAL,
	},
	{
		expected:
			"Showing 50 of 1234 assets that match “pnl”. Add more of the ID or name, or choose a type, to narrow the search.",
		receivedCount: RECEIVED,
		search: TEXT_ONLY,
		totalItems: MANY,
	},
];

describe("getSearchSummaryText", () => {
	it.each(CASES)("$expected", ({ expected, ...options }) => {
		expect(getSearchSummaryText(options)).toBe(expected);
	});
});
