import { createContext, useContext } from 'react';

export interface AboutContextType {
	version: string;
	commit_hash: string;
	platform: string;
	path: string;
	lang: string;
	enable_workers: boolean;
	uptime: string;
	statup_time: string;
	gin_mode: string;
	gin_version: string;
	go_version: string;
	node_version: string;
}

export const AboutContext = createContext<AboutContextType | undefined>(undefined);
export const AboutContextProvider = AboutContext.Provider;

export function useAbout() {
	const context = useContext(AboutContext);
	if (!context) {
		throw new Error('useAbout must be used within an AboutProvider');
	}
	return context;
}
