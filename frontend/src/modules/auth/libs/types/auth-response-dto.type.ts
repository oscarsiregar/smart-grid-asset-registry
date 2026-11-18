import { type UserDto } from "~/modules/users/users.js";

type AuthResponseDto = {
	token: string;
	user: UserDto;
};

export { type AuthResponseDto };
