import { useCallback, useEffect, useId, useRef } from "react";

import { Button } from "~/libs/components/button/button.js";
import { ButtonVariant, ControlSize, IconName } from "~/libs/enums/enums.js";

import { PanelLabel } from "./libs/enums/enums.js";
import styles from "./styles.module.css";

type Properties = {
	children: React.ReactNode;
	/** A one-line subtitle under the title, also the dialog's description. */
	description?: string | undefined;
	isOpen: boolean;
	onClose: () => void;
	title: string;
};

/**
 * A modal panel on the native `<dialog>` element (constitution, principle V).
 * `showModal()` makes the rest of the page inert, which traps focus, and the
 * browser fires `cancel` on Escape. The panel never closes itself: Escape and
 * the Close button call `onClose`, and the owner decides (for example by
 * removing a URL search parameter). Focus returns to the element that was
 * focused when the panel opened. Children are only mounted while open, so
 * every opening starts from a fresh state.
 */
const Panel: React.FC<Properties> = ({
	children,
	description,
	isOpen,
	onClose,
	title,
}: Properties) => {
	const dialogReference = useRef<HTMLDialogElement>(null);
	const openerReference = useRef<Element | null>(null);
	const titleId = useId();
	const descriptionId = useId();

	useEffect(() => {
		const dialog = dialogReference.current;

		if (!dialog) {
			return;
		}

		if (isOpen && !dialog.open) {
			openerReference.current = dialog.ownerDocument.activeElement;
			dialog.showModal();

			return;
		}

		if (!isOpen && dialog.open) {
			dialog.close();

			if (openerReference.current instanceof HTMLElement) {
				openerReference.current.focus();
			}

			openerReference.current = null;
		}
	}, [isOpen]);

	const handleCancel = useCallback(
		(event: React.SyntheticEvent<HTMLDialogElement>): void => {
			event.preventDefault();
			onClose();
		},
		[onClose],
	);

	return (
		<dialog
			aria-describedby={description === undefined ? undefined : descriptionId}
			aria-labelledby={titleId}
			className={styles["panel"]}
			onCancel={handleCancel}
			ref={dialogReference}
		>
			<header className={styles["header"]}>
				<div className={styles["heading"]}>
					<h2 className={styles["title"]} id={titleId}>
						{title}
					</h2>
					{description !== undefined && (
						<p className={styles["description"]} id={descriptionId}>
							{description}
						</p>
					)}
				</div>
				<Button
					iconName={IconName.CLOSE}
					isLabelHidden
					label={PanelLabel.CLOSE}
					onClick={onClose}
					size={ControlSize.SM}
					type="button"
					variant={ButtonVariant.SECONDARY}
				/>
			</header>
			<div className={styles["body"]}>{isOpen && children}</div>
		</dialog>
	);
};

export { Panel };
