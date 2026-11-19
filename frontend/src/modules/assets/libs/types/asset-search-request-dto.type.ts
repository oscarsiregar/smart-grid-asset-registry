import { type z } from "zod";

import { type assetSearch } from "../validation-schemas/asset-search.validation-schema.js";

type AssetSearchRequestDto = z.infer<typeof assetSearch>;

export { type AssetSearchRequestDto };
