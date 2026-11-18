import { useCallback } from "react";

import { Button } from "~/libs/components/button/button.js";
import { ButtonVariant, ControlSize, IconName } from "~/libs/enums/enums.js";
import { useDeleteNoteMutation } from "~/modules/notes/notes-api.js";
import { type NoteDto } from "~/modules/notes/notes.js";

import { NotesPageMessage } from "../../../../libs/enums/enums.js";
import { formatNoteDate } from "../../../../libs/helpers/helpers.js";
import styles from "./styles.module.css";

type Properties = {
	note: NoteDto;
};

const NoteItem: React.FC<Properties> = ({
	note: { content, createdAt, id, title },
}: Properties) => {
	const [deleteNote, { isLoading }] = useDeleteNoteMutation();

	const handleDeleteClick = useCallback((): void => {
		void deleteNote(id);
	}, [deleteNote, id]);

	return (
		<li className={styles["item"]}>
			<div className={styles["body"]}>
				<h3 className={styles["title"]}>{title}</h3>
				{content && <p className={styles["content"]}>{content}</p>}
				<time className={styles["date"]} dateTime={createdAt}>
					{formatNoteDate(createdAt)}
				</time>
			</div>
			<Button
				iconName={IconName.TRASH_2}
				isLoading={isLoading}
				label={NotesPageMessage.DELETE_LABEL}
				onClick={handleDeleteClick}
				size={ControlSize.SM}
				type="button"
				variant={ButtonVariant.DANGER}
			/>
		</li>
	);
};

export { NoteItem };
