/** How an outcome is styled. Colour is never the only signal: every tone
 * comes with an icon and the outcome text (constitution, principle V). */
const OutcomeTone = {
	DANGER: "danger",
	INFO: "info",
	SUCCESS: "success",
	WARNING: "warning",
} as const;

export { OutcomeTone };
