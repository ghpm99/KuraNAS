import { fireEvent, render, screen } from '@testing-library/react';
import React from 'react';
import VideoContent from './videoContent';

const mockUseVideoPlaylists = jest.fn();
const mockUseAllVideoFiles = jest.fn();
const mockUseVideoPlaylistDetail = jest.fn();
const mockUseQuery = jest.fn();
const mockUseMutation = jest.fn();
const mockUseQueryClient = jest.fn();
const mockNavigate = jest.fn();
const mockSetSearchParams = jest.fn();
const mockSetSearch = { value: '' };
const mockInvalidateQueries = jest.fn();

const mockAddVideoToPlaylist = jest.fn();
const mockGetVideoPlaylistById = jest.fn();
const mockRemoveVideoFromPlaylist = jest.fn();
const mockReorderVideoPlaylist = jest.fn();
const mockUpdateVideoPlaylistName = jest.fn();
const mockGetVideoPlaybackState = jest.fn();

jest.mock('@/components/hooks/useVideos/useVideos', () => ({
	useVideoPlaylists: () => mockUseVideoPlaylists(),
	useAllVideoFiles: () => mockUseAllVideoFiles(),
	useVideoPlaylistDetail: (...args: any[]) => mockUseVideoPlaylistDetail(...args),
}));

jest.mock('@/service/videoPlayback', () => ({
	addVideoToPlaylist: (...args: any[]) => mockAddVideoToPlaylist(...args),
	getVideoPlaylistById: (...args: any[]) => mockGetVideoPlaylistById(...args),
	removeVideoFromPlaylist: (...args: any[]) => mockRemoveVideoFromPlaylist(...args),
	reorderVideoPlaylist: (...args: any[]) => mockReorderVideoPlaylist(...args),
	updateVideoPlaylistName: (...args: any[]) => mockUpdateVideoPlaylistName(...args),
	getVideoPlaybackState: (...args: any[]) => mockGetVideoPlaybackState(...args),
}));

jest.mock('@tanstack/react-query', () => ({
	useQuery: (...args: any[]) => mockUseQuery(...args),
	useMutation: (...args: any[]) => mockUseMutation(...args),
	useQueryClient: () => mockUseQueryClient(),
}));

jest.mock('react-router-dom', () => ({
	useNavigate: () => mockNavigate,
	useSearchParams: () => [new URLSearchParams(mockSetSearch.value), mockSetSearchParams],
}));

jest.mock('@/service/apiUrl', () => ({
	getApiV1BaseUrl: () => 'http://localhost:8000/v1',
}));
jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string, params?: Record<string, string | number>) => {
			const map: Record<string, string> = {
				VIDEO_LOADING_VIDEOS: 'Carregando videos...',
				VIDEO_CONTINUE_WATCHING: 'Continuar assistindo',
				VIDEO_ALL: 'Todos',
				VIDEO_SEARCH_PLACEHOLDER: 'Buscar video por nome, pasta ou formato',
				VIDEO_PLAY: 'Reproduzir',
				VIDEO_ADD: 'Adicionar',
				VIDEO_ADD_SUCCESS: 'Video adicionado a playlist com sucesso.',
				VIDEO_NO_RECENT_PLAYLISTS: 'Nenhuma playlist com reproducao recente.',
				VIDEO_ADD_ERROR: 'Nao foi possivel adicionar o video a playlist.',
				VIDEO_ALREADY_ADDED: 'Ja adicionado',
				VIDEO_EDIT_PLAYLIST: 'Editar playlist',
				VIDEO_SAVE_NAME: 'Salvar nome',
				VIDEO_BACK_TO_VIDEOS: 'Voltar para videos',
				VIDEO_DISPLAY_NAME_PLACEHOLDER: 'Nome de exibicao',
				VIDEO_LOADING_PLAYLIST: 'Carregando playlist...',
			};
			if (key === 'VIDEO_PREVIEW_ALT') return `${params?.name ?? ''} preview`.trim();
			if (key === 'VIDEO_PLAYLIST_ITEM_COUNT') return `${params?.count ?? 0} videos`;
			if (key === 'VIDEO_PLAYLIST_META') return `${params?.count ?? 0} videos nesta playlist`;
			return map[key] ?? key;
		},
	}),
}));

const playlist = {
	id: 1,
	name: 'Playlist One',
	classification: 'series',
	item_count: 1,
	cover_video_id: 30,
	last_played_at: '2026-03-04T10:00:00.000Z',
	items: [
		{
			id: 50,
			order_index: 0,
			source_kind: 'manual',
			video: {
				id: 30,
				name: 'ep-1.mp4',
				format: 'mp4',
				path: '/videos/ep-1.mp4',
				parent_path: '/videos',
				size: 10,
			},
		},
	],
};

const detailPlaylist = {
	...playlist,
	items: [
		{
			id: 50,
			order_index: 0,
			source_kind: 'manual',
			video: { id: 30, name: 'ep-1.mp4', format: 'mp4', path: '/videos/ep-1.mp4', parent_path: '/videos', size: 10 },
		},
		{
			id: 51,
			order_index: 1,
			source_kind: 'auto',
			video: { id: 31, name: 'ep-2.mkv', format: 'mkv', path: '/videos/ep-2.mkv', parent_path: '/videos', size: 12 },
		},
	],
};

beforeEach(() => {
	jest.clearAllMocks();
	mockSetSearch.value = '';
	mockUseQueryClient.mockReturnValue({ invalidateQueries: mockInvalidateQueries });
	mockUseVideoPlaylists.mockReturnValue({ data: [playlist], isLoading: false });
	mockUseAllVideoFiles.mockReturnValue({
		data: [{ id: 30, name: 'ep-1.mp4', parent_path: '/videos', format: 'mp4' }],
		isLoading: false,
	});
	mockUseVideoPlaylistDetail.mockReturnValue({ data: playlist, isLoading: false });
	mockGetVideoPlaylistById.mockResolvedValue(detailPlaylist);
	mockGetVideoPlaybackState.mockResolvedValue({ playback_state: { playlist_id: 1, video_id: 30 } });
	mockAddVideoToPlaylist.mockResolvedValue({});
	mockRemoveVideoFromPlaylist.mockResolvedValue({});
	mockReorderVideoPlaylist.mockResolvedValue({});
	mockUpdateVideoPlaylistName.mockResolvedValue({});

	mockUseQuery.mockImplementation((options: any) => {
		const [key] = options.queryKey;
		if (options.enabled !== false) {
			options.queryFn?.();
		}
		if (key === 'video-playback-state') {
			return { data: { playback_state: { playlist_id: 1, video_id: 30 } } };
		}
		if (key === 'video-playlist-membership') {
			return { data: {} };
		}
		return { data: undefined };
	});

	mockUseMutation.mockImplementation((options: any) => ({
		mutate: (variables: any) => {
			options.mutationFn?.(variables);
			options.onSuccess?.();
		},
		isPending: false,
	}));
});

describe('components/videos/videoContent', () => {
	it('renders loading state', () => {
		mockUseVideoPlaylists.mockReturnValue({ data: [], isLoading: true });
		render(<VideoContent />);
		expect(screen.getByText('Carregando videos...')).toBeInTheDocument();
	});

	it('renders catalog and handles add/play/search actions', async () => {
		render(<VideoContent />);

		expect(screen.getByText('Continuar assistindo')).toBeInTheDocument();
		expect(screen.getAllByText('Playlist One').length).toBeGreaterThan(0);
		expect(screen.getByText('Todos')).toBeInTheDocument();

		fireEvent.change(screen.getByPlaceholderText('Buscar video por nome, pasta ou formato'), {
			target: { value: 'no-match' },
		});
		expect(screen.queryByText('ep-1.mp4')).not.toBeInTheDocument();
		fireEvent.change(screen.getByPlaceholderText('Buscar video por nome, pasta ou formato'), {
			target: { value: 'ep-1' },
		});
		expect(screen.getByText('ep-1.mp4')).toBeInTheDocument();

		fireEvent.click(screen.getAllByRole('button', { name: /Reproduzir/i })[0]);
		expect(mockNavigate).toHaveBeenCalled();
		fireEvent.click(screen.getAllByRole('button', { name: /Playlist One/ })[0]);
		expect(mockSetSearchParams).toHaveBeenCalled();

		fireEvent.change(screen.getByRole('combobox'), { target: { value: '1' } });

		fireEvent.click(screen.getByRole('button', { name: /Adicionar/i }));
		expect(mockAddVideoToPlaylist).toHaveBeenCalledWith(1, 30);
		expect(await screen.findByText('Video adicionado a playlist com sucesso.')).toBeInTheDocument();
	});

	it('renders empty continue section when there is no playback history', () => {
		mockUseVideoPlaylists.mockReturnValue({
			data: [{ ...playlist, id: 2, last_played_at: null, classification: 'unknown-class' }],
			isLoading: false,
		});
		render(<VideoContent />);
		expect(screen.getByText('Nenhuma playlist com reproducao recente.')).toBeInTheDocument();
		expect(screen.getByText('unknown-class')).toBeInTheDocument();
	});

	it('renders playlist card without cover and keeps play button disabled', () => {
		mockUseVideoPlaylists.mockReturnValue({
			data: [{ ...playlist, id: 9, name: 'No Cover', classification: 'movie', cover_video_id: null, items: [], last_played_at: null }],
			isLoading: false,
		});
		mockUseAllVideoFiles.mockReturnValue({
			data: [{ id: 0, name: 'invalid-id.mp4', parent_path: '/videos', format: 'mp4' }],
			isLoading: false,
		});
		mockUseQuery.mockImplementation((options: any) => {
			const [key] = options.queryKey;
			if (key === 'video-playback-state') return { data: undefined };
			if (key === 'video-playlist-membership') return { data: {} };
			return { data: undefined };
		});

		render(<VideoContent />);
		const playButtons = screen.getAllByRole('button', { name: /Reproduzir/i });
		expect(playButtons[0]).toBeDisabled();
		fireEvent.click(playButtons[1]);
		expect(mockNavigate).not.toHaveBeenCalledWith('/video/0', expect.anything());
		expect(screen.getAllByText('No Cover').length).toBeGreaterThan(0);
	});

	it('sorts continue playlists and handles add-to-playlist mutation error', async () => {
		mockUseVideoPlaylists.mockReturnValue({
			data: [
				{ ...playlist, id: 2, name: 'Older', last_played_at: '2026-03-03T10:00:00.000Z' },
				{ ...playlist, id: 3, name: 'Newer', last_played_at: '2026-03-04T12:00:00.000Z' },
			],
			isLoading: false,
		});
		mockUseMutation.mockImplementation((options: any) => ({
			mutate: (variables: any) => {
				options.mutationFn?.(variables);
				options.onError?.();
			},
			isPending: false,
		}));

		render(<VideoContent />);
		const headers = screen.getAllByText(/Older|Newer/).map((el) => el.textContent);
		expect(headers[0]).toBe('Newer');
		fireEvent.click(screen.getAllByRole('button', { name: /Adicionar/i })[0]);
		expect(await screen.findByText('Nao foi possivel adicionar o video a playlist.')).toBeInTheDocument();
	});

	it('renders selected playlist detail branch', () => {
		mockSetSearch.value = 'playlist=playlist-one';
		mockUseVideoPlaylistDetail.mockReturnValue({ data: detailPlaylist, isLoading: false });
		render(<VideoContent />);

		expect(screen.getByText('Editar playlist')).toBeInTheDocument();
		expect(screen.getByText('Salvar nome')).toBeInTheDocument();
		fireEvent.click(screen.getByRole('button', { name: /Voltar para videos/i }));
		expect(mockSetSearchParams).toHaveBeenCalled();
	});

	it('handles add-to-playlist error and already-added branch', async () => {
		mockUseQuery.mockImplementation((options: any) => {
			const [key] = options.queryKey;
			if (options.enabled !== false) options.queryFn?.();
			if (key === 'video-playback-state') {
				return { data: { playback_state: { playlist_id: 1, video_id: 30 } } };
			}
			if (key === 'video-playlist-membership') {
				return { data: { 1: new Set([30]) } };
			}
			return { data: undefined };
		});
		mockUseMutation.mockImplementation((options: any) => ({
			mutate: (variables: any) => {
				options.mutationFn?.(variables);
				options.onError?.();
			},
			isPending: false,
		}));

		render(<VideoContent />);
		expect(screen.getByRole('button', { name: /Ja adicionado/i })).toBeDisabled();
		fireEvent.click(screen.getAllByRole('button', { name: /Reproduzir/i })[0]);
		expect(mockNavigate).toHaveBeenCalled();
		await screen.findByText('Continuar assistindo');
	});

	it('renders selected playlist loading branch', () => {
		mockSetSearch.value = 'playlist=playlist-one';
		mockUseVideoPlaylistDetail.mockReturnValue({ data: undefined, isLoading: true });
		render(<VideoContent />);
		expect(screen.getByText('Carregando playlist...')).toBeInTheDocument();
	});

	it('runs playlist detail actions: rename reorder remove and navigate', async () => {
		mockSetSearch.value = 'playlist=playlist-one';
		mockUseVideoPlaylistDetail.mockReturnValue({ data: detailPlaylist, isLoading: false });
		mockUseMutation.mockImplementation((options: any) => ({
			mutate: (...args: any[]) => {
				options.mutationFn?.(...args);
				options.onSuccess?.(...args);
			},
			isPending: false,
		}));

		render(<VideoContent />);

		fireEvent.change(screen.getByPlaceholderText('Nome de exibicao'), { target: { value: '' } });
		expect(screen.getByRole('button', { name: 'Salvar nome' })).toBeDisabled();

		fireEvent.change(screen.getByPlaceholderText('Nome de exibicao'), { target: { value: 'Renamed' } });
		fireEvent.click(screen.getByRole('button', { name: 'Salvar nome' }));
		expect(mockUpdateVideoPlaylistName).toHaveBeenCalledWith(1, 'Renamed');

		fireEvent.click(screen.getAllByRole('button').find((b) => b.className.includes('iconBtnDanger')) as HTMLElement);
		expect(mockRemoveVideoFromPlaylist).toHaveBeenCalledWith(1, 30);

		const moveDownButton = screen
			.getAllByRole('button')
			.find((b) => b.className.includes('iconBtn') && !b.className.includes('Danger') && !b.disabled) as HTMLElement;
		fireEvent.click(moveDownButton);
		expect(mockReorderVideoPlaylist).toHaveBeenCalled();

		fireEvent.click(screen.getByText('ep-1.mp4'));
		expect(mockNavigate).toHaveBeenCalled();
		expect(mockInvalidateQueries).toHaveBeenCalled();
		await screen.findByText('Editar playlist');
	});
});
