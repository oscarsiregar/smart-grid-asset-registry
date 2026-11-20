/** The positions 0 … count − 1, for rendering a fixed number of placeholders. */
const createPositions = (count: number): number[] =>
	Array.from({ length: count }, (_, position) => position);

export { createPositions };
