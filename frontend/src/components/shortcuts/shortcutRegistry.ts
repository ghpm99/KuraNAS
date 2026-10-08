import { createContext, useContext, useEffect, useId } from 'react';
import type { ShortcutDefinition } from './shortcutDefinition';

export type ShortcutRegistry = {
    pageShortcuts: ShortcutDefinition[];
    registerPageShortcuts: (ownerId: string, definitions: ShortcutDefinition[]) => () => void;
};

const emptyRegistry: ShortcutRegistry = {
    pageShortcuts: [],
    registerPageShortcuts: () => () => undefined,
};

export const ShortcutRegistryContext = createContext<ShortcutRegistry>(emptyRegistry);

export const usePageShortcuts = (): ShortcutDefinition[] =>
    useContext(ShortcutRegistryContext).pageShortcuts;

export const useRegisterPageShortcuts = (definitions: ShortcutDefinition[], isEnabled = true) => {
    const ownerId = useId();
    const { registerPageShortcuts } = useContext(ShortcutRegistryContext);

    useEffect(() => {
        if (!isEnabled) return undefined;
        return registerPageShortcuts(ownerId, definitions);
    }, [definitions, isEnabled, ownerId, registerPageShortcuts]);
};
