import { render } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import SettingsProvider from '@/components/providers/settingsProvider';
import { GlobalMusicProvider } from './GlobalMusicProvider';

// Resilience-first (no-mock render): mount the raw provider with only the
// minimal real context it structurally needs (QueryClient + SettingsProvider)
// and NO backend mock. Every API call fails against a non-existent server and
// it must not throw. This is the exact component that crashed when the backend
// returned a partial settings payload and `settings.players` was undefined.
describe('GlobalMusicProvider (no-mock render)', () => {
    it('renders without throwing when the backend is unavailable', () => {
        const queryClient = new QueryClient({
            defaultOptions: { queries: { retry: false } },
        });

        expect(() =>
            render(
                <QueryClientProvider client={queryClient}>
                    <SettingsProvider>
                        <GlobalMusicProvider>
                            <div>musica</div>
                        </GlobalMusicProvider>
                    </SettingsProvider>
                </QueryClientProvider>
            )
        ).not.toThrow();
    });
});
