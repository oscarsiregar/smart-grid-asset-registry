import { render, screen } from "@testing-library/react";
import { Provider } from "react-redux";
import { afterEach, describe, expect, it, vi } from "vitest";

import { HTTPCode } from "~/libs/enums/enums.js";
import { config } from "~/libs/modules/config/config.js";
import { createStore } from "~/libs/modules/store/store.module.js";
import {
	type AssetNodeDto,
	AssetType,
	OperationalStatus,
} from "~/modules/assets/assets.js";

import { RegisterSummary } from "./register-summary.js";

const MANY = 13;
const ONE = 1;
const NONE = 0;

const createSubstations = (count: number): AssetNodeDto[] =>
	Array.from({ length: count }, (_, index) => ({
		assetId: `SS-${String(index)}`,
		assetName: `Substation ${String(index)}`,
		assetType: AssetType.SUBSTATION,
		childCount: 0,
		operationalStatus: OperationalStatus.IN_SERVICE,
	}));

const stubTopLevel = (count: number) => {
	const fetchMock = vi.fn(() =>
		Promise.resolve(
			Response.json(
				{
					code: HTTPCode.OK,
					data: { items: createSubstations(count), totalItems: count },
					message: "ok",
					status: "success",
				},
				{ status: HTTPCode.OK },
			),
		),
	);
	vi.stubGlobal("fetch", fetchMock);

	return fetchMock;
};

const renderSummary = (): void => {
	render(
		<Provider store={createStore(config)}>
			<RegisterSummary />
		</Provider>,
	);
};

afterEach(() => {
	vi.unstubAllGlobals();
});

describe("RegisterSummary", () => {
	it("counts the stored substations from the top-level list", async () => {
		const fetchMock = stubTopLevel(MANY);
		renderSummary();

		expect(await screen.findByText("substations")).toHaveTextContent(
			"13 substations",
		);
		expect(screen.getByText("Register")).toBeInTheDocument();
		const [[request]] = fetchMock.mock.calls as unknown as [[Request]];
		expect(request.url).toBe("http://localhost/api/assets");
		expect(fetchMock).toHaveBeenCalledOnce();
	});

	it("says one substation in the singular", async () => {
		stubTopLevel(ONE);
		renderSummary();

		expect(await screen.findByText("substation")).toHaveTextContent(
			"1 substation",
		);
	});

	it("shows a placeholder for the number while the top level loads", () => {
		vi.stubGlobal(
			"fetch",
			vi.fn(
				() =>
					new Promise<Response>(() => {
						// Never settles.
					}),
			),
		);
		renderSummary();

		expect(screen.getByText("Register")).toBeInTheDocument();
		expect(document.querySelector("[data-shape]")).toBeInTheDocument();
		expect(screen.queryByText(/substation/)).not.toBeInTheDocument();
	});

	it("shows nothing for an empty registry", async () => {
		const fetchMock = stubTopLevel(NONE);
		const { container } = render(
			<Provider store={createStore(config)}>
				<RegisterSummary />
			</Provider>,
		);

		await vi.waitFor(() => {
			expect(fetchMock).toHaveBeenCalledOnce();
		});
		await vi.waitFor(() => {
			expect(container).toBeEmptyDOMElement();
		});
	});
});
