import { PageContainer } from "~/libs/components/page-container/page-container.js";
import { PageIntro } from "~/libs/components/page-intro/page-intro.js";

import { NoteCreateForm } from "./components/note-create-form/note-create-form.js";
import { NotesList } from "./components/notes-list/notes-list.js";
import { NotesPageMessage } from "./libs/enums/enums.js";
import styles from "./styles.module.css";

const Notes: React.FC = () => (
	<PageContainer>
		<div className={styles["content"]}>
			<PageIntro
				description={NotesPageMessage.DESCRIPTION}
				label={NotesPageMessage.KICKER}
				title={NotesPageMessage.TITLE}
			/>
			<NoteCreateForm />
			<NotesList />
		</div>
	</PageContainer>
);

export { Notes };
