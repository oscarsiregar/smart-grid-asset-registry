import { type FieldPath } from "react-hook-form";

import { type SignInRequestDto } from "~/modules/auth/auth.js";

const SIGN_IN_FIELDS: FieldPath<SignInRequestDto>[] = ["email", "password"];

export { SIGN_IN_FIELDS };
