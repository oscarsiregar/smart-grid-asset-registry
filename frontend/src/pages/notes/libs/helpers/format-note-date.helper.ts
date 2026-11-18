const noteDateFormatter = new Intl.DateTimeFormat(undefined, {
	dateStyle: "medium",
	timeStyle: "short",
});

const formatNoteDate = (date: string): string => {
	return noteDateFormatter.format(new Date(date));
};

export { formatNoteDate };
