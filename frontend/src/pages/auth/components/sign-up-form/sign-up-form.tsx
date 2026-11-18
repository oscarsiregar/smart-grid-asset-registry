import { useCallback } from "react";
import { Navigate } from "react-router-dom";

import { Button } from "~/libs/components/button/button.js";
import { FormAlert } from "~/libs/components/form-alert/form-alert.js";
import { Input } from "~/libs/components/input/input.js";
import { Link } from "~/libs/components/link/link.js";
import { AppRoute, ControlSize, InputType } from "~/libs/enums/enums.js";
import { useAppForm } from "~/libs/hooks/use-app-form/use-app-form.hook.js";
import { useServerFormErrors } from "~/libs/hooks/use-server-form-errors/use-server-form-errors.hook.js";
import { checkIsToastedError } from "~/libs/modules/api/libs/helpers/check-is-toasted-error.helper.js";
import { useSignUpMutation } from "~/modules/auth/auth-api.js";
import {
	type SignUpRequestDto,
	signUpValidationSchema,
} from "~/modules/auth/auth.js";

import styles from "../../styles.module.css";
import {
	DEFAULT_SIGN_UP_PAYLOAD,
	SIGN_UP_FIELDS,
} from "./libs/constants/constants.js";

const SignUpForm: React.FC = () => {
	const [signUp, { data, error, isError, isLoading }] = useSignUpMutation();

	const { clearErrors, control, handleSubmit, setError } =
		useAppForm<SignUpRequestDto>({
			defaultValues: DEFAULT_SIGN_UP_PAYLOAD,
			validationSchema: signUpValidationSchema,
		});

	const { hasFieldErrors } = useServerFormErrors({
		clearErrors,
		error,
		fields: SIGN_UP_FIELDS,
		setError,
	});
	const generalError =
		hasFieldErrors || checkIsToastedError(error) ? undefined : error;

	const handleFormSubmit = useCallback(
		(event: React.BaseSyntheticEvent): void => {
			void handleSubmit((payload: SignUpRequestDto) => void signUp(payload))(
				event,
			);
		},
		[handleSubmit, signUp],
	);

	if (!isError && Boolean(data?.user)) {
		return <Navigate replace to={AppRoute.DASHBOARD} />;
	}

	return (
		<>
			<h1 className={styles["heading"]}>Sign up</h1>
			<FormAlert error={generalError} />
			<form className={styles["form"]} noValidate onSubmit={handleFormSubmit}>
				<div className={styles["fields"]}>
					<Input
						autoComplete="email"
						control={control}
						isRequired
						label="Email"
						name="email"
						placeholder="Enter your email"
						type={InputType.EMAIL}
					/>
					<Input
						autoComplete="nickname"
						control={control}
						isRequired
						label="Nickname"
						name="nickname"
						placeholder="Choose a nickname"
					/>
					<Input
						autoComplete="new-password"
						control={control}
						isRequired
						label="Password"
						name="password"
						placeholder="Create a password"
						type={InputType.PASSWORD}
					/>
				</div>
				<Button
					isLoading={isLoading}
					label="Create account"
					size={ControlSize.LG}
					type="submit"
				/>
			</form>
			<p className={styles["footer"]}>
				Already have an account?{" "}
				<Link className={styles["inline-link"]} to={AppRoute.SIGN_IN}>
					Sign in
				</Link>
			</p>
		</>
	);
};

export { SignUpForm };
