import { useCallback, useMemo, useState, type ReactNode } from 'react';
import type { ShortcutDefinition } from './shortcutDefinition';
import { ShortcutRegistryContext, type ShortcutRegistry } from './shortcutRegistry';

const ShortcutRegistryProvider = ({ children }: { children: ReactNode }) => {
    const [definitionsByOwnerId, setDefinitionsByOwnerId] = useState<
        Record<string, ShortcutDefinition[]>
    >({});

    const registerPageShortcuts = useCallback(
        (ownerId: string, definitions: ShortcutDefinition[]) => {
            setDefinitionsByOwnerId((current) => ({ ...current, [ownerId]: definitions }));
            return () =>
                setDefinitionsByOwnerId((current) =>
                    Object.fromEntries(
                        Object.entries(current).filter(([registeredOwnerId]) => registeredOwnerId !== ownerId)
                    )
                );
        },
        []
    );

    const registry = useMemo<ShortcutRegistry>(
        () => ({
            pageShortcuts: Object.values(definitionsByOwnerId).flat(),
            registerPageShortcuts,
        }),
        [definitionsByOwnerId, registerPageShortcuts]
    );

    return (
        <ShortcutRegistryContext.Provider value={registry}>
            {children}
        </ShortcutRegistryContext.Provider>
    );
};


export default ShortcutRegistryProvider;
