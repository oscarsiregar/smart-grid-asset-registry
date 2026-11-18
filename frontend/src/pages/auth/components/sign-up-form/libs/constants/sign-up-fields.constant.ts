import { type FieldPath } from "react-hook-form";

import { type SignUpRequestDto } from "~/modules/auth/auth.js";

const SIGN_UP_FIELDS: FieldPath<SignUpRequestDto>[] = [
	"email",
	"nickname",
	"password",
];

export { SIGN_UP_FIELDS };
