import { type z } from "zod";

import { type signUp } from "../validation-schemas/sign-up.validation-schema.js";

type SignUpRequestDto = z.infer<typeof signUp>;

export { type SignUpRequestDto };
