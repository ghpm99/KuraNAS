import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import AddToPlaylistMenu, { AddToPlaylistButton } from './AddToPlaylistMenu';

const mockUseQuery = jest.fn();
const mockUseMutation = jest.fn();
const mockUseQueryClient = jest.fn();
const mockEnqueueSnackbar = jest.fn();
const mockGetPlaylists = jest.fn();
const mockAddTrackToPlaylist = jest.fn();
const mockCreatePlaylist = jest.fn();

jest.mock('@tanstack/react-query', () => ({
    useQuery: (...args: any[]) => mockUseQuery(...args),
    useMutation: (...args: any[]) => mockUseMutation(...args),
    useQueryClient: () => mockUseQueryClient(),
}));

jest.mock('notistack', () => ({
    useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

jest.mock('@/service/playlist', () => ({
    getPlaylists: (...args: any[]) => mockGetPlaylists(...args),
    addTrackToPlaylist: (...args: any[]) => mockAddTrackToPlaylist(...args),
    createPlaylist: (...args: any[]) => mockCreatePlaylist(...args),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

describe('components/music/AddToPlaylistMenu', () => {
    const invalidateQueries = jest.fn();
    const onClose = jest.fn();
    const anchor = document.createElement('button');

    beforeEach(() => {
        jest.clearAllMocks();
        mockUseQueryClient.mockReturnValue({ invalidateQueries });
        mockGetPlaylists.mockResolvedValue({
            items: [
                {
                    id: 10,
                    name: 'My Playlist',
                    is_system: false,
                    is_auto: false,
                    kind: 'manual',
                    source_key: '',
                },
            ],
        });
        mockAddTrackToPlaylist.mockResolvedValue({});
        mockCreatePlaylist.mockResolvedValue({ id: 99, name: 'Roadtrip' });
        mockUseQuery.mockImplementation((options: any) => {
            if (options.enabled !== false) options.queryFn?.();
            return {
                data: {
                    items: [
                        {
                            id: 10,
                            name: 'My Playlist',
                            is_system: false,
                            is_auto: false,
                            kind: 'manual',
                            source_key: '',
                        },
                        {
                            id: 11,
                            name: 'System Playlist',
                            is_system: true,
                            is_auto: true,
                            kind: 'automatic',
                            source_key: 'favorites',
                        },
                    ],
                },
                isLoading: false,
            };
        });
        mockUseMutation.mockImplementation((options: any) => ({
            mutate: (variables: any) => {
                options.mutationFn?.(variables);
                options.onSuccess?.();
            },
            isPending: false,
        }));
    });

    it('shows loading state in menu', () => {
        mockUseQuery.mockReturnValue({ data: undefined, isLoading: true });
        render(<AddToPlaylistMenu fileId={5} anchorEl={anchor} onClose={onClose} />);
        expect(screen.getByText('LOADING')).toBeInTheDocument();
    });

    it('adds track to selected playlist', () => {
        render(<AddToPlaylistMenu fileId={5} anchorEl={anchor} onClose={onClose} />);
        fireEvent.click(screen.getByText('My Playlist'));
        expect(mockAddTrackToPlaylist).toHaveBeenCalledWith(10, 5);
        expect(invalidateQueries).toHaveBeenCalled();
        expect(mockEnqueueSnackbar).toHaveBeenCalledWith('MUSIC_TRACK_ADDED', {
            variant: 'success',
        });
        expect(onClose).toHaveBeenCalled();
    });

    it('creates new playlist and adds track', () => {
        render(<AddToPlaylistMenu fileId={5} anchorEl={anchor} onClose={onClose} />);
        fireEvent.click(screen.getByText('MUSIC_NEW_PLAYLIST'));
        expect(screen.getByText('MUSIC_CREATE_PLAYLIST_ADD')).toBeInTheDocument();

        fireEvent.change(screen.getByLabelText('MUSIC_PLAYLIST_NAME'), {
            target: { value: 'Roadtrip' },
        });
        fireEvent.click(screen.getByRole('button', { name: 'ACTION_CREATE_ADD' }));

        return waitFor(() => {
            expect(mockCreatePlaylist).toHaveBeenCalledWith({ name: 'Roadtrip' });
            expect(mockAddTrackToPlaylist).toHaveBeenCalledWith(99, 5);
            expect(invalidateQueries).toHaveBeenCalled();
            expect(mockEnqueueSnackbar).toHaveBeenCalledWith('MUSIC_PLAYLIST_CREATED_ADDED', {
                variant: 'success',
            });
            expect(onClose).toHaveBeenCalled();
        });
    });

    it('shows warning when add mutation fails and opens button wrapper', () => {
        mockUseMutation
            .mockImplementationOnce((options: any) => ({
                mutate: () => options.onError?.(),
                isPending: false,
            }))
            .mockImplementationOnce((options: any) => ({
                mutate: () => options.onSuccess?.(),
                isPending: false,
            }));

        render(<AddToPlaylistMenu fileId={5} anchorEl={anchor} onClose={onClose} />);
        fireEvent.click(screen.getByText('My Playlist'));
        expect(mockEnqueueSnackbar).toHaveBeenCalledWith('MUSIC_TRACK_ADD_FAILED', {
            variant: 'warning',
        });

        render(<AddToPlaylistButton fileId={6} />);
        fireEvent.click(screen.getAllByRole('button')[0]!);
    });

    it('shows empty playlists entry and handles create failure', () => {
        mockUseQuery.mockImplementation((options: any) => {
            if (options.enabled !== false) options.queryFn?.();
            return { data: { items: [] }, isLoading: false };
        });
        mockUseMutation.mockImplementation((options: any) => ({
            mutate: () => options.onError?.(),
            isPending: false,
        }));

        render(<AddToPlaylistMenu fileId={7} anchorEl={anchor} onClose={onClose} />);
        expect(screen.getByText('MUSIC_NO_PLAYLISTS')).toBeInTheDocument();
        fireEvent.click(screen.getByText('MUSIC_NEW_PLAYLIST'));
        fireEvent.change(screen.getByLabelText('MUSIC_PLAYLIST_NAME'), {
            target: { value: 'Fails' },
        });
        fireEvent.click(screen.getByRole('button', { name: 'ACTION_CREATE_ADD' }));
        expect(mockEnqueueSnackbar).toHaveBeenCalledWith('MUSIC_PLAYLIST_CREATE_FAILED', {
            variant: 'error',
        });
    });
});
