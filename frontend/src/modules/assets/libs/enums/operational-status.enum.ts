/** Operational statuses as the API sends them; `null` means none recorded. */
const OperationalStatus = {
	IN_SERVICE: "IN_SERVICE",
	MAINTENANCE: "MAINTENANCE",
	OUT_OF_SERVICE: "OUT_OF_SERVICE",
} as const;

export { OperationalStatus };
