import { type z } from "zod";

import { type signIn } from "../validation-schemas/sign-in.validation-schema.js";

type SignInRequestDto = z.infer<typeof signIn>;

export { type SignInRequestDto };
