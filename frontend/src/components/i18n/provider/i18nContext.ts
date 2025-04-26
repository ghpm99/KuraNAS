import { createContext, useContext } from 'react';

export type I18nContextType = {
	t: (key: string, options?: Record<string, string>) => string;
};

const I18nContext = createContext<I18nContextType | undefined>(undefined);

export const I18nContextProvider = I18nContext.Provider;

export const useI18n = () => {
	const context = useContext(I18nContext);
	if (!context) {
		throw new Error('useI18n must be used within an I18nProvider');
	}
	return context;
};

export default useI18n;
