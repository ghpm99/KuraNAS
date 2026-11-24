import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import type { ReactNode } from 'react';
import { PlaylistsProvider, usePlaylistsProvider } from './playlistsProvider';
import { createEmptyPlaylistRules } from './playlistRuleCriteria';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn(), put: jest.fn(), delete: jest.fn() },
}));

const mockEnqueueSnackbar = jest.fn();

jest.mock('notistack', () => ({
    useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock; put: jest.Mock };

const rockRules = { ...createEmptyPlaylistRules(), match: 'all' as const, genres: ['rock'] };

const renderProvider = (initialEntry = '/') => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const wrapper = ({ children }: { children: ReactNode }) => (
        <QueryClientProvider client={client}>
            <MemoryRouter initialEntries={[initialEntry]}>
                <PlaylistsProvider>{children}</PlaylistsProvider>
            </MemoryRouter>
        </QueryClientProvider>
    );
    return renderHook(() => usePlaylistsProvider(), { wrapper });
};

describe('features/music/playlistsProvider rules (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockImplementation((url: string) => {
            if (url === '/music/playlists/system') return Promise.resolve({ data: [] });
            return Promise.resolve({
                data: {
                    items: [{ id: 3, name: 'Rock', is_system: false, kind: 'rule' }],
                    pagination: { has_next: false },
                },
            });
        });
        mockedApi.post.mockResolvedValue({ data: { id: 9 } });
        mockedApi.put.mockResolvedValue({ data: { id: 3 } });
    });

    it('creates a manual playlist without rules in the body', async () => {
        const { result } = renderProvider();
        act(() => result.current.openCreateDialog());
        act(() => result.current.setNewName('Mix'));

        act(() => result.current.submitCreatePlaylist());

        await waitFor(() =>
            expect(mockedApi.post).toHaveBeenCalledWith('/music/playlists/', {
                name: 'Mix',
                description: '',
            })
        );
    });

    it('creates a rule playlist posting the exact rules', async () => {
        const { result } = renderProvider();
        act(() => result.current.openCreateDialog());
        act(() => result.current.setNewKind('rule'));
        act(() => result.current.setNewName('Rock'));
        act(() => result.current.setNewDescription('loud'));
        act(() => result.current.setNewRules(rockRules));

        act(() => result.current.submitCreatePlaylist());

        await waitFor(() =>
            expect(mockedApi.post).toHaveBeenCalledWith('/music/playlists/', {
                name: 'Rock',
                description: 'loud',
                rules: rockRules,
            })
        );
        await waitFor(() => expect(result.current.createOpen).toBe(false));
        expect(result.current.newKind).toBe('manual');
        expect(result.current.newRules).toEqual(createEmptyPlaylistRules());
    });

    it('shows the server message verbatim and keeps the offending field on a 400', async () => {
        mockedApi.post.mockRejectedValue({
            response: { data: { error: 'Informe ao menos um critério', field: 'rules' } },
        });
        const { result } = renderProvider();
        act(() => result.current.openCreateRuleDialog());
        act(() => result.current.setNewName('Vazia'));

        act(() => result.current.submitCreatePlaylist());

        await waitFor(() =>
            expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Informe ao menos um critério', {
                variant: 'error',
            })
        );
        expect(result.current.createErrorField).toBe('rules');
        expect(result.current.createOpen).toBe(true);

        act(() => result.current.setNewRules(rockRules));
        expect(result.current.createErrorField).toBeNull();
    });

    it('openCreateRuleDialog opens the dialog in rule mode prefilled with the initial rules', () => {
        const { result } = renderProvider();

        act(() => result.current.openCreateRuleDialog({ genres: ['metal'] }));

        expect(result.current.createOpen).toBe(true);
        expect(result.current.newKind).toBe('rule');
        expect(result.current.newRules).toEqual({
            ...createEmptyPlaylistRules(),
            genres: ['metal'],
        });
    });

    it('opens the create dialog prefilled from the create_rule_genre URL param', () => {
        const { result } = renderProvider('/?create_rule_genre=heavy-metal');

        expect(result.current.createOpen).toBe(true);
        expect(result.current.newKind).toBe('rule');
        expect(result.current.newRules.genres).toEqual(['heavy-metal']);
    });

    it('ignores the URL param when absent', () => {
        const { result } = renderProvider();

        expect(result.current.createOpen).toBe(false);
        expect(result.current.newKind).toBe('manual');
    });

    it('closing the dialog drops the URL param so it does not reopen', async () => {
        const { result } = renderProvider('/?create_rule_genre=metal');

        act(() => result.current.closeCreateDialog());

        expect(result.current.createOpen).toBe(false);
    });

    it('updates a rule playlist PUTting name, description and rules', async () => {
        const { result } = renderProvider();
        await waitFor(() => expect(result.current.playlists.some((p) => p.id === 3)).toBe(true));
        act(() => result.current.selectPlaylist({ id: 3, name: 'Rock' } as never));
        await waitFor(() => expect(result.current.selectedPlaylist?.id).toBe(3));
        const onSaved = jest.fn();

        act(() => result.current.renameSelectedPlaylist('Rock', 'loud', onSaved, rockRules));

        await waitFor(() =>
            expect(mockedApi.put).toHaveBeenCalledWith('/music/playlists/3', {
                name: 'Rock',
                description: 'loud',
                rules: rockRules,
            })
        );
        await waitFor(() => expect(onSaved).toHaveBeenCalled());
    });

    it('keeps the rules out of the PUT body for a manual rename and exposes the field on 400', async () => {
        const { result } = renderProvider();
        await waitFor(() => expect(result.current.playlists.some((p) => p.id === 3)).toBe(true));
        act(() => result.current.selectPlaylist({ id: 3, name: 'Rock' } as never));
        await waitFor(() => expect(result.current.selectedPlaylist?.id).toBe(3));

        act(() => result.current.renameSelectedPlaylist('Mix', 'd'));
        await waitFor(() =>
            expect(mockedApi.put).toHaveBeenCalledWith('/music/playlists/3', {
                name: 'Mix',
                description: 'd',
            })
        );

        mockedApi.put.mockRejectedValue({
            response: { data: { error: 'Ano inválido', field: 'year_from' } },
        });
        act(() => result.current.renameSelectedPlaylist('Mix', 'd', undefined, rockRules));

        await waitFor(() => expect(result.current.renameErrorField).toBe('year_from'));
        expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Ano inválido', { variant: 'error' });
    });
});
