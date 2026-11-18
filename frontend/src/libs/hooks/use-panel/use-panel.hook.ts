import { useCallback } from "react";
import { useSearchParams } from "react-router-dom";

import { type PanelName, SearchParameter } from "~/libs/enums/enums.js";
import { type ValueOf } from "~/libs/types/types.js";

type PanelControls = {
	closePanel: () => void;
	isOpen: boolean;
	openPanel: () => void;
};

/**
 * Keeps a panel's open state in the URL (`?panel=<name>`), so refresh, deep
 * links and the back button work (constitution, principle V). Opening adds a
 * history entry, so Back closes the panel; closing replaces the current entry.
 * Other search parameters are kept.
 */
const usePanel = (name: ValueOf<typeof PanelName>): PanelControls => {
	const [searchParameters, setSearchParameters] = useSearchParams();

	const isOpen = searchParameters.get(SearchParameter.PANEL) === name;

	const openPanel = useCallback((): void => {
		setSearchParameters((current) => {
			const next = new URLSearchParams(current);
			next.set(SearchParameter.PANEL, name);

			return next;
		});
	}, [name, setSearchParameters]);

	const closePanel = useCallback((): void => {
		setSearchParameters(
			(current) => {
				const next = new URLSearchParams(current);
				next.delete(SearchParameter.PANEL);

				return next;
			},
			{ replace: true },
		);
	}, [setSearchParameters]);

	return { closePanel, isOpen, openPanel };
};

export { usePanel };
