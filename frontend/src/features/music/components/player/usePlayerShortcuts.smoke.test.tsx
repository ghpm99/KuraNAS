import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render } from '@testing-library/react';
import I18nProvider from '@/components/i18n/provider';
import SettingsProvider from '@/components/providers/settingsProvider';
import { GlobalMusicProvider } from '@/features/music/providers/GlobalMusicProvider';
import { usePlayerShortcuts } from './usePlayerShortcuts';

const PlayerShortcutsProbe = () => {
    usePlayerShortcuts();
    return null;
};

describe('usePlayerShortcuts (no-mock render)', () => {
    it('mounts without throwing when the backend is unavailable', () => {
        const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });

        expect(() =>
            render(
                <QueryClientProvider client={queryClient}>
                    <I18nProvider>
                        <SettingsProvider>
                            <GlobalMusicProvider>
                                <PlayerShortcutsProbe />
                            </GlobalMusicProvider>
                        </SettingsProvider>
                    </I18nProvider>
                </QueryClientProvider>
            )
        ).not.toThrow();
    });
});
