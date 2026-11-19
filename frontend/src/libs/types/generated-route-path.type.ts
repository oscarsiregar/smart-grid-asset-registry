/**
 * An address built from an `AppRoute` with its parameters filled in and
 * encoded (`/assets/SS%2F1`). Only `generateRoutePath` creates one, so a
 * plain string cannot be passed where a generated address is expected.
 */
type GeneratedRoutePath = string & {
	readonly __brand: "GeneratedRoutePath";
};

export { type GeneratedRoutePath };
