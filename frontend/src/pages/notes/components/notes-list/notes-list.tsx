import { useCallback } from "react";

import { Button } from "~/libs/components/button/button.js";
import { Loader } from "~/libs/components/loader/loader.js";
import { EMPTY_LENGTH } from "~/libs/constants/constants.js";
import { ButtonVariant, ControlSize } from "~/libs/enums/enums.js";
import { useGetNotesQuery } from "~/modules/notes/notes-api.js";

import { NotesPageMessage } from "../../libs/enums/enums.js";
import { NoteItem } from "./components/note-item/note-item.js";
import styles from "./styles.module.css";

const NotesList: React.FC = () => {
	const {
		data: notes,
		isError,
		isLoading,
		refetch,
	} = useGetNotesQuery(undefined);

	const handleRetryClick = useCallback((): void => {
		void refetch();
	}, [refetch]);

	if (isLoading) {
		return <Loader />;
	}

	if (isError) {
		return (
			<div className={styles["state"]} role="alert">
				<p className={styles["state-message"]}>{NotesPageMessage.LOAD_ERROR}</p>
				<Button
					label={NotesPageMessage.RETRY_LABEL}
					onClick={handleRetryClick}
					size={ControlSize.SM}
					type="button"
					variant={ButtonVariant.SECONDARY}
				/>
			</div>
		);
	}

	if (!notes || notes.length === EMPTY_LENGTH) {
		return <p className={styles["state-message"]}>{NotesPageMessage.EMPTY}</p>;
	}

	return (
		<section className={styles["section"]}>
			<h2 className={styles["title"]}>{NotesPageMessage.LIST_TITLE}</h2>
			<ul className={styles["list"]}>
				{notes.map((note) => (
					<NoteItem key={note.id} note={note} />
				))}
			</ul>
		</section>
	);
};

export { NotesList };
