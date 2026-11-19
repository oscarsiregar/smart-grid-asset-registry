import { describe, expect, it } from "vitest";

import { ImportMode } from "../enums/enums.js";
import { createImportFormData } from "./create-import-form-data.helper.js";

describe("createImportFormData", () => {
	it("puts the file and the mode into the multipart parts the API expects", () => {
		const file = new File(["asset_id\n"], "grid.csv", { type: "text/csv" });

		const formData = createImportFormData({ file, mode: ImportMode.STRICT });

		expect(formData.get("file")).toBe(file);
		expect(formData.get("mode")).toBe("STRICT");
		expect(formData.keys().toArray()).toEqual(["file", "mode"]);
	});
});
