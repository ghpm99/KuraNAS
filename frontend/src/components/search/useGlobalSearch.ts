import { createContext, useContext } from 'react';

export interface GlobalSearchContextValue {
    openSearch: () => void;
    shortcut: string;
}

export const GlobalSearchContext = createContext<GlobalSearchContextValue | undefined>(undefined);

export const useGlobalSearch = () => {
    const context = useContext(GlobalSearchContext);
    if (!context) {
        throw new Error('useGlobalSearch must be used within a GlobalSearchProvider');
    }
    return context;
};

export default useGlobalSearch;
