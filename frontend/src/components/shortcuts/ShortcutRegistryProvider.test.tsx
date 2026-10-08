import { render, screen } from '@testing-library/react';
import ShortcutRegistryProvider from './ShortcutRegistryProvider';
import { usePageShortcuts, useRegisterPageShortcuts } from './shortcutRegistry';
import type { ShortcutDefinition } from './shortcutDefinition';

const openDefinitions: ShortcutDefinition[] = [{ keyLabels: ['Enter'], descriptionKey: 'OPEN' }];

const Registrant = ({
    definitions,
    isEnabled = true,
}: {
    definitions: ShortcutDefinition[];
    isEnabled?: boolean;
}) => {
    useRegisterPageShortcuts(definitions, isEnabled);
    return null;
};

const RegisteredList = () => (
    <ul>
        {usePageShortcuts().map((shortcut) => (
            <li key={shortcut.descriptionKey}>{shortcut.descriptionKey}</li>
        ))}
    </ul>
);

describe('shortcut registry', () => {
    it('registers without a provider without crashing and exposes nothing', () => {
        render(
            <>
                <Registrant definitions={openDefinitions} />
                <RegisteredList />
            </>
        );

        expect(screen.queryAllByRole('listitem')).toHaveLength(0);
    });

    it('lists the shortcuts of mounted registrants and drops them on unmount', () => {
        const { rerender } = render(
            <ShortcutRegistryProvider>
                <Registrant definitions={openDefinitions} />
                <RegisteredList />
            </ShortcutRegistryProvider>
        );
        expect(screen.getByText('OPEN')).toBeInTheDocument();

        rerender(
            <ShortcutRegistryProvider>
                <RegisteredList />
            </ShortcutRegistryProvider>
        );
        expect(screen.queryByText('OPEN')).not.toBeInTheDocument();
    });

    it('ignores disabled registrants', () => {
        render(
            <ShortcutRegistryProvider>
                <Registrant definitions={openDefinitions} isEnabled={false} />
                <RegisteredList />
            </ShortcutRegistryProvider>
        );

        expect(screen.queryByText('OPEN')).not.toBeInTheDocument();
    });
});
