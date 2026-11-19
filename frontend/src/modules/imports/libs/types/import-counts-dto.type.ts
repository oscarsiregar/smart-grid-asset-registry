/** `total` = `imported` + `unchanged` + `rejected`. */
type ImportCountsDto = {
	imported: number;
	rejected: number;
	total: number;
	unchanged: number;
	warningRows: number;
};

export { type ImportCountsDto };
