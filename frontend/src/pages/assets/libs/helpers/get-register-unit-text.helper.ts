import { SINGLE_SUBSTATION_COUNT } from "../constants/constants.js";
import { AssetsPageMessage } from "../enums/enums.js";

/** The word after the register count: "substations", or "substation" for one (spec 005 FR-009). */
const getRegisterUnitText = (count: number): string =>
	count === SINGLE_SUBSTATION_COUNT
		? AssetsPageMessage.REGISTER_UNIT_ONE
		: AssetsPageMessage.REGISTER_UNIT;

export { getRegisterUnitText };
