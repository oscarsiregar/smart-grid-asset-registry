/** Assets beneath a substation on every level, by type; a type with none is 0. */
type DescendantCountsDto = {
	lvBoards: number;
	switchboardPanels: number;
	switchboards: number;
	transformers: number;
};

export { type DescendantCountsDto };
