/** Roving tab index of the tree: one node is reachable with Tab, the rest are skipped. */
const TabIndex = {
	FOCUSABLE: 0,
	SKIPPED: -1,
} as const;

export { TabIndex };
