import ActivePageListener from '@/components/activePageListener';
import { UIProvider } from '@/components/providers/uiProvider';
import type { ReactNode } from 'react';
import { AppShell as AppShellComponent } from './AppShell';

const AppShell = ({ children }: { children: ReactNode }) => {
    return (
        <UIProvider>
            <ActivePageListener />
            <AppShellComponent>{children}</AppShellComponent>
        </UIProvider>
    );
};

export default AppShell;
