import { useState } from 'react';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { apiBase } from '@/service';
import { PlaylistRules } from '@/types/playlist';
import { createEmptyPlaylistRules } from '@/features/music/providers/playlistsProvider/playlistRuleCriteria';
import PlaylistRuleForm from './PlaylistRuleForm';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, options?: Record<string, string>) =>
            options ? `${key}:${options.count}` : key,
    }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock };

const serverPage = (items: unknown[]) => ({ data: { items, pagination: { has_next: false } } });

function ControlledForm({ initialRules }: { initialRules: PlaylistRules }) {
    const [rules, setRules] = useState(initialRules);
    return <PlaylistRuleForm rules={rules} onChange={setRules} />;
}

const renderForm = (initialRules = createEmptyPlaylistRules()) => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={client}>
            <ControlledForm initialRules={initialRules} />
        </QueryClientProvider>
    );
};

const previewCalls = () =>
    mockedApi.post.mock.calls.filter(([url]) => url === '/music/playlists/rules/preview');

describe('features/music PlaylistRuleForm (seam)', () => {
    beforeEach(() => {
        jest.useFakeTimers();
        jest.clearAllMocks();
        mockedApi.get.mockImplementation((url: string) => {
            if (url === '/music/library/genres') {
                return Promise.resolve(
                    serverPage([{ key: 'rock', genre: 'Rock', track_count: 3 }])
                );
            }
            return Promise.resolve(serverPage([]));
        });
        mockedApi.post.mockResolvedValue({ data: { track_count: 7 } });
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    it('does not call the preview endpoint while the rule has no criterion', async () => {
        renderForm();

        await act(async () => {
            jest.advanceTimersByTime(1000);
        });

        expect(previewCalls()).toHaveLength(0);
    });

    it('debounces 300 ms and posts the exact rule to the preview endpoint', async () => {
        renderForm();

        fireEvent.change(screen.getByLabelText('MUSIC_PLAYLIST_RULE_YEAR_FROM'), {
            target: { value: '1990' },
        });
        expect(screen.getByRole('status')).toHaveTextContent('MUSIC_PLAYLIST_RULE_PREVIEW_LOADING');

        await act(async () => {
            jest.advanceTimersByTime(299);
        });
        expect(previewCalls()).toHaveLength(0);

        fireEvent.change(screen.getByLabelText('MUSIC_PLAYLIST_RULE_YEAR_TO'), {
            target: { value: '1999' },
        });
        await act(async () => {
            jest.advanceTimersByTime(299);
        });
        expect(previewCalls()).toHaveLength(0);

        await act(async () => {
            jest.advanceTimersByTime(1);
        });

        await waitFor(() => expect(previewCalls()).toHaveLength(1));
        expect(previewCalls()[0]).toEqual([
            '/music/playlists/rules/preview',
            { ...createEmptyPlaylistRules(), year_from: 1990, year_to: 1999 },
        ]);
        await waitFor(() =>
            expect(screen.getByRole('status')).toHaveTextContent(
                'MUSIC_PLAYLIST_RULE_PREVIEW_COUNT:7'
            )
        );
    });

    it('shows the server error message verbatim when the preview is rejected', async () => {
        mockedApi.post.mockRejectedValue({ response: { data: { error: 'Regra inválida' } } });
        renderForm({ ...createEmptyPlaylistRules(), favorites_only: true });

        await act(async () => {
            jest.advanceTimersByTime(300);
        });

        await waitFor(() => expect(screen.getByRole('status')).toHaveTextContent('Regra inválida'));
    });

    it('falls back to the translated message when the preview fails without a server message', async () => {
        mockedApi.post.mockRejectedValue(new Error('network'));
        renderForm({ ...createEmptyPlaylistRules(), favorites_only: true });

        await act(async () => {
            jest.advanceTimersByTime(300);
        });

        await waitFor(() =>
            expect(screen.getByRole('status')).toHaveTextContent(
                'MUSIC_PLAYLIST_RULE_PREVIEW_FAILED'
            )
        );
    });

    it('loads the genre options from the server and keeps the selected key in the rule', async () => {
        renderForm();
        await waitFor(() =>
            expect(mockedApi.get).toHaveBeenCalledWith('/music/library/genres', {
                params: { page: 1, page_size: 500 },
            })
        );

        const genresInput = screen.getByLabelText('MUSIC_PLAYLIST_RULE_GENRES');
        fireEvent.mouseDown(genresInput);
        fireEvent.click(await screen.findByText('Rock'));

        await act(async () => {
            jest.advanceTimersByTime(300);
        });
        await waitFor(() =>
            expect(previewCalls()[0]).toEqual([
                '/music/playlists/rules/preview',
                { ...createEmptyPlaylistRules(), genres: ['rock'] },
            ])
        );
    });
});
