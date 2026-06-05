import { createContext, useContext } from 'react';

export type I18nContextType = {
    t: (key: string, options?: Record<string, string>) => string;
};

const I18nContext = createContext<I18nContextType | undefined>(undefined);

export const I18nContextProvider = I18nContext.Provider;

// Resilient fallback: outside an I18nProvider (or while it is unavailable) we
// echo the key back instead of throwing. Consumers like the crash-screen
// ErrorBoundary must never themselves crash because of a missing provider;
// the worst-case degraded UX is showing raw keys, which is loud and obvious.
const fallbackContext: I18nContextType = {
    t: (key: string) => key,
};

export const useI18n = (): I18nContextType => {
    return useContext(I18nContext) ?? fallbackContext;
};

export default useI18n;
