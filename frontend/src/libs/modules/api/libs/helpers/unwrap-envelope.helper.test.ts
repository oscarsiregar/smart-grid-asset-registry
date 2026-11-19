import { describe, expect, it } from "vitest";

import { unwrapEnvelope } from "./unwrap-envelope.helper.js";

describe("unwrapEnvelope", () => {
	it("returns the data of a success envelope", () => {
		const data = { outcome: "VALIDATED_ONLY" };

		expect(
			unwrapEnvelope({ code: 200, data, message: "ok", status: "success" }),
		).toBe(data);
	});

	it("keeps a null data", () => {
		expect(
			unwrapEnvelope({
				code: 200,
				data: null,
				message: "ok",
				status: "success",
			}),
		).toBeNull();
	});

	it.each([
		["a body without the envelope", { id: 1 }],
		["a string", "plain"],
		["null", null],
		[
			"an error envelope",
			{ code: 422, data: null, message: "x", status: "error" },
		],
	])("leaves %s unchanged", (_label, body) => {
		expect(unwrapEnvelope(body)).toBe(body);
	});
});
