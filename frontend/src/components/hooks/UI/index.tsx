import { ReactNode, useState } from 'react';
import { pages, UIContext } from './uiContext';

export function UIProvider({ children }: { children: ReactNode }) {
	const [activePage, setActivePage] = useState<pages>('files');

	const value = {
		activePage,
		setActivePage,
	};

	console.log('UIProvider activePage:', activePage);

	return <UIContext.Provider value={value}>{children}</UIContext.Provider>;
}
