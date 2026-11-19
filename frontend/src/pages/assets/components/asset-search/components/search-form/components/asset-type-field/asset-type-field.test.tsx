import { fireEvent, render, screen } from "@testing-library/react";
import { useWatch } from "react-hook-form";
import { describe, expect, it } from "vitest";

import { useAppForm } from "~/libs/hooks/use-app-form/use-app-form.hook.js";
import {
	type AssetSearchRequestDto,
	AssetType,
} from "~/modules/assets/assets.js";

import { AssetTypeField } from "./asset-type-field.js";

type HarnessProperties = {
	assetType: AssetSearchRequestDto["assetType"];
};

// The field inside a form, with the form's value shown as JSON.
const Harness: React.FC<HarnessProperties> = ({
	assetType,
}: HarnessProperties) => {
	const { control } = useAppForm<AssetSearchRequestDto>({
		defaultValues: { assetType, q: "" },
	});
	const value = useWatch({ control, name: "assetType" });

	return (
		<>
			<AssetTypeField control={control} name="assetType" />
			<output data-testid="value">{JSON.stringify(value)}</output>
		</>
	);
};

const getSelect = (): HTMLElement =>
	screen.getByRole("combobox", { name: "Asset type" });

describe("AssetTypeField", () => {
	it("offers All types and the five types in words", () => {
		render(<Harness assetType={null} />);

		expect(
			screen.getAllByRole("option").map((option) => option.textContent),
		).toEqual([
			"All types",
			"Substation",
			"Transformer",
			"LV board",
			"Switchboard",
			"Switchboard panel",
		]);
		expect(getSelect()).toHaveDisplayValue("All types");
	});

	it("sets the chosen type", () => {
		render(<Harness assetType={null} />);

		fireEvent.change(getSelect(), { target: { value: AssetType.TRANSFORMER } });

		expect(screen.getByTestId("value")).toHaveTextContent(
			JSON.stringify(AssetType.TRANSFORMER),
		);
	});

	it("turns All types back into null", () => {
		render(<Harness assetType={AssetType.TRANSFORMER} />);
		expect(getSelect()).toHaveDisplayValue("Transformer");

		fireEvent.change(getSelect(), { target: { value: "ALL" } });

		expect(screen.getByTestId("value")).toHaveTextContent("null");
	});
});
