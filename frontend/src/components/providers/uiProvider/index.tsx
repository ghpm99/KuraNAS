import { ReactNode, useMemo, useState } from 'react';
import { pages, UIContext } from './uiContext';

export function UIProvider({ children }: { children: ReactNode }) {
    const [activePage, setActivePage] = useState<pages>('unknown');

    const value = useMemo(
        () => ({
            activePage,
            setActivePage,
        }),
        [activePage, setActivePage]
    );

    return <UIContext.Provider value={value}>{children}</UIContext.Provider>;
}
