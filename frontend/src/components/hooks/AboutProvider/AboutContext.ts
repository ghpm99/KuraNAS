import { createContext, useContext } from 'react';

export interface AboutContextType {
	version: string;
	commit_hash: string;
	platform: string;
	path: string;
	lang: string;
	enable_workers: boolean;
	statup_time: string; // formato: "2006-01-02 15:04:05"
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
