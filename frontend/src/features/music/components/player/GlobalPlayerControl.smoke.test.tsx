import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render } from '@testing-library/react';
import I18nProvider from '@/components/i18n/provider';
import SettingsProvider from '@/components/providers/settingsProvider';
import { GlobalMusicProvider } from '@/features/music/providers/GlobalMusicProvider';
import GlobalPlayerControl from './GlobalPlayerControl';

describe('GlobalPlayerControl (no-mock render)', () => {
    it('renders without throwing when the backend is unavailable', () => {
        const queryClient = new QueryClient({
            defaultOptions: { queries: { retry: false } },
        });

        expect(() =>
            render(
                <QueryClientProvider client={queryClient}>
                    <I18nProvider>
                        <SettingsProvider>
                            <GlobalMusicProvider>
                                <GlobalPlayerControl />
                            </GlobalMusicProvider>
                        </SettingsProvider>
                    </I18nProvider>
                </QueryClientProvider>
            )
        ).not.toThrow();
    });
});
