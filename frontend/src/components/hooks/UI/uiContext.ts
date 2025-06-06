import { createContext, useContext } from 'react';

export type pages = 'files' | 'activity' | 'analytics' | 'about';

interface UIContextType {
	activePage: pages;
	setActivePage: (page: pages) => void;
}

export const UIContext = createContext<UIContextType | undefined>(undefined);

export function useUI() {
	const context = useContext(UIContext);
	if (context === undefined) {
		throw new Error('useUI must be used within a UIProvider');
	}
	return context;
}
