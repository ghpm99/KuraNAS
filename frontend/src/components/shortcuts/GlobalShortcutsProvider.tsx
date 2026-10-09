import { useCallback, useEffect, useState, type ReactNode } from 'react';
import { subscribeToShortcutsHelp } from '@/components/layout/appCommandEvents';
import GlobalShortcutsDialog from './GlobalShortcutsDialog';
import { globalShortcutDefinitions } from './globalShortcutDefinitions';
import ShortcutRegistryProvider from './ShortcutRegistryProvider';
import { usePageShortcuts } from './shortcutRegistry';
import { useGlobalShortcuts } from './useGlobalShortcuts';

const GlobalShortcutsHost = ({ children }: { children: ReactNode }) => {
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const pageShortcuts = usePageShortcuts();
    const openDialog = useCallback(() => setIsDialogOpen(true), []);
    const closeDialog = useCallback(() => setIsDialogOpen(false), []);
    useGlobalShortcuts(openDialog);
    useEffect(() => subscribeToShortcutsHelp(openDialog), [openDialog]);

    return (
        <>
            {children}
            <GlobalShortcutsDialog
                isOpen={isDialogOpen}
                onClose={closeDialog}
                globalShortcuts={globalShortcutDefinitions}
                pageShortcuts={pageShortcuts}
            />
        </>
    );
};

const GlobalShortcutsProvider = ({ children }: { children: ReactNode }) => (
    <ShortcutRegistryProvider>
        <GlobalShortcutsHost>{children}</GlobalShortcutsHost>
    </ShortcutRegistryProvider>
);

export default GlobalShortcutsProvider;
