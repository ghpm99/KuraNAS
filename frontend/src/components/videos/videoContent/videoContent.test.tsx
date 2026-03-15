import { fireEvent, render, screen } from '@testing-library/react';
import VideoContent from './videoContent';

const mockUseQuery = jest.fn();
const mockUseMutation = jest.fn();
const mockUseQueryClient = jest.fn();
const mockNavigate = jest.fn();
const mockInvalidateQueries = jest.fn();
const mockLocation = { pathname: '/videos', search: '' };

const mockGetVideoPlaylists = jest.fn();
const mockGetAllVideoFiles = jest.fn();
const mockGetVideoHomeCatalog = jest.fn();
const mockAddVideoToPlaylist = jest.fn();
const mockGetVideoPlaylistById = jest.fn();
const mockRemoveVideoFromPlaylist = jest.fn();
const mockReorderVideoPlaylist = jest.fn();
const mockUpdateVideoPlaylistName = jest.fn();
const mockGetVideoPlaybackState = jest.fn();

jest.mock('@/service/videoPlayback', () => ({
	getVideoPlaylists: (...args: any[]) => mockGetVideoPlaylists(...args),
	getAllVideoFiles: (...args: any[]) => mockGetAllVideoFiles(...args),
	getVideoHomeCatalog: (...args: any[]) => mockGetVideoHomeCatalog(...args),
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
	Link: ({ children, to }: any) => <a href={to}>{children}</a>,
	useLocation: () => mockLocation,
	useNavigate: () => mockNavigate,
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
				VIDEO_LOADING_PLAYLIST: 'Carregando playlist...',
				VIDEO_SECTION_CONTINUE: 'Continuar assistindo',
				VIDEO_SECTION_CONTINUE_DESCRIPTION: 'Playlists e series onde voce parou recentemente.',
				VIDEO_SECTION_SERIES: 'Series',
				VIDEO_SECTION_SERIES_DESCRIPTION: 'Colecoes automaticas para series e animes com sequencia clara.',
				VIDEO_SECTION_MOVIES: 'Filmes',
				VIDEO_SECTION_MOVIES_DESCRIPTION: 'Filmes organizados como biblioteca dedicada, sem misturar com outros contextos.',
				VIDEO_SECTION_PERSONAL: 'Pessoais',
				VIDEO_SECTION_PERSONAL_DESCRIPTION: 'Gravacoes e videos pessoais agrupados com contexto proprio.',
				VIDEO_SECTION_CLIPS: 'Clipes',
				VIDEO_SECTION_CLIPS_DESCRIPTION: 'Clipes curtos e videos de programa em uma trilha rapida de consumo.',
				VIDEO_SECTION_FOLDERS: 'Pastas',
				VIDEO_SECTION_FOLDERS_DESCRIPTION: 'Entrada por pasta para quando a organizacao fisica importa mais que a classificacao.',
				VIDEO_SECTION_SERIES_EMPTY: 'Nenhuma serie classificada ainda.',
				VIDEO_SECTION_MOVIES_EMPTY: 'Nenhum filme classificado ainda.',
				VIDEO_SECTION_PERSONAL_EMPTY: 'Nenhum video pessoal classificado ainda.',
				VIDEO_SECTION_CLIPS_EMPTY: 'Nenhum clipe classificado ainda.',
				VIDEO_SECTION_FOLDERS_EMPTY: 'Nenhuma pasta de video agrupada ainda.',
				VIDEO_OPEN_SECTION: 'Abrir secao',
				VIDEO_HOME_RECENT: 'Adicionados recentemente',
				VIDEO_HOME_RECENT_DESCRIPTION: 'Ultimos videos detectados no catalogo para abrir direto do dominio.',
				VIDEO_STATUS_NOT_STARTED: 'Nao iniciado',
				VIDEO_STATUS_IN_PROGRESS: 'Em andamento',
				VIDEO_STATUS_COMPLETED: 'Concluido',
				VIDEO_SEARCH_PLACEHOLDER: 'Buscar video por nome, pasta ou formato',
				VIDEO_PLAY: 'Reproduzir',
				VIDEO_ADD: 'Adicionar',
				VIDEO_ADD_SUCCESS: 'Video adicionado a playlist com sucesso.',
				VIDEO_ADD_ERROR: 'Nao foi possivel adicionar o video a playlist.',
				VIDEO_ALREADY_ADDED: 'Ja adicionado',
				VIDEO_ALL: 'Todos',
				VIDEO_ALL_DESC: 'Todos os videos do sistema com acao rapida para adicionar a playlists.',
				VIDEO_NO_RECENT_PLAYLISTS: 'Nenhuma playlist com reproducao recente.',
				VIDEO_EDIT_PLAYLIST: 'Editar playlist',
				VIDEO_SAVE_NAME: 'Salvar nome',
				VIDEO_BACK_TO_VIDEOS: 'Voltar para videos',
				VIDEO_DISPLAY_NAME_PLACEHOLDER: 'Nome de exibicao',
				VIDEO_CONTINUE_BADGE_RESUME: 'Retomar',
				VIDEO_SOURCE_MANUAL: 'manual',
				VIDEO_SOURCE_AUTO: 'auto',
				VIDEO_DETAIL_COLLECTION_META: '{{count}} itens',
				VIDEO_DETAIL_COLLECTION_TOTAL: 'Total',
				VIDEO_DETAIL_COLLECTION_COMPLETED: 'Concluidos',
				VIDEO_DETAIL_COLLECTION_PENDING: 'Pendentes',
				VIDEO_DETAIL_COLLECTION_ITEMS: 'Itens da colecao',
				VIDEO_DETAIL_COLLECTION_ITEMS_DESCRIPTION: 'Continue do ponto certo ou abra qualquer item do contexto.',
				VIDEO_DETAIL_RESUME_ACTION: 'Retomar agora',
				VIDEO_DETAIL_SERIES_EYEBROW: 'Serie',
				VIDEO_DETAIL_SEASON_LABEL: 'Temporada {{season}}',
				VIDEO_DETAIL_SEASON_DESCRIPTION: '{{count}} episodios organizados em sequencia.',
			};
			if (key === 'VIDEO_PREVIEW_ALT') return `${params?.name ?? ''} preview`.trim();
			if (key === 'VIDEO_PLAYLIST_ITEM_COUNT') return `${params?.count ?? 0} videos`;
			if (key === 'VIDEO_PLAYLIST_META') return `${params?.count ?? 0} videos nesta playlist`;
			if (key === 'VIDEO_DETAIL_SERIES_PROGRESS') {
				return `${params?.completed ?? 0} de ${params?.count ?? 0} concluidos`;
			}
			if (key === 'VIDEO_DETAIL_COLLECTION_META') return `${params?.count ?? 0} itens`;
			if (key === 'VIDEO_DETAIL_SEASON_LABEL') return `Temporada ${params?.season ?? ''}`.trim();
			if (key === 'VIDEO_DETAIL_SEASON_DESCRIPTION') return `${params?.count ?? 0} episodios organizados em sequencia.`;
			return map[key] ?? key;
		},
	}),
}));

const playlist = {
	id: 1,
	name: 'Playlist One',
	classification: 'series',
	type: 'series',
	item_count: 1,
	cover_video_id: 30,
	last_played_at: '2026-03-04T10:00:00.000Z',
	items: [
		{
			id: 50,
			order_index: 0,
			source_kind: 'manual',
			status: 'in_progress',
			progress_pct: 55,
			video: { id: 30, name: 'Show S01E01.mp4', format: 'mp4', path: '/videos/ep-1.mp4', parent_path: '/videos', size: 10 },
		},
	],
};

const folderPlaylist = {
	...playlist,
	id: 2,
	name: 'Folder Playlist',
	type: 'folder',
	classification: 'personal',
	cover_video_id: 31,
	last_played_at: null,
};

const clipPlaylist = {
	...playlist,
	id: 3,
	name: 'Clip Playlist',
	type: 'custom',
	classification: 'clip',
	cover_video_id: 32,
	last_played_at: null,
};

const moviePlaylist = {
	...playlist,
	id: 4,
	name: 'Movie Playlist',
	type: 'movie',
	classification: 'movie',
	cover_video_id: 33,
	last_played_at: null,
};

const personalPlaylist = {
	...playlist,
	id: 5,
	name: 'Personal Playlist',
	type: 'custom',
	classification: 'personal',
	cover_video_id: 34,
	last_played_at: null,
};

const detailPlaylist = {
	...playlist,
	items: [
		{
			id: 50,
			order_index: 0,
			source_kind: 'manual',
			status: 'completed',
			progress_pct: 100,
			video: { id: 30, name: 'Show S01E01.mp4', format: 'mp4', path: '/videos/ep-1.mp4', parent_path: '/videos', size: 10 },
		},
		{
			id: 51,
			order_index: 1,
			source_kind: 'auto',
			status: 'in_progress',
			progress_pct: 45,
			video: { id: 31, name: 'Show S01E02.mkv', format: 'mkv', path: '/videos/ep-2.mkv', parent_path: '/videos', size: 12 },
		},
	],
};

const folderDetailPlaylist = {
	...folderPlaylist,
	items: [
		{
			id: 52,
			order_index: 0,
			source_kind: 'manual',
			status: 'not_started',
			progress_pct: 0,
			video: { id: 31, name: 'Home Clip.mp4', format: 'mp4', path: '/videos/home-clip.mp4', parent_path: '/videos/home', size: 10 },
		},
	],
};

const clipDetailPlaylist = {
	...clipPlaylist,
	items: [
		{
			id: 60,
			order_index: 0,
			source_kind: 'auto',
			status: 'not_started',
			progress_pct: 0,
			video: { id: 32, name: 'Quick Clip.mp4', format: 'mp4', path: '/clips/quick-clip.mp4', parent_path: '/clips', size: 10 },
		},
	],
};

let playlistsData: any[] = [];
let allVideosData: any[] = [];
let homeCatalogData: any = undefined;
let playbackData: any = undefined;
let membershipData: Record<number, Set<number>> = {};
let selectedPlaylistData: any = undefined;
let selectedPlaylistLoading = false;
let mutationShouldError = false;

beforeEach(() => {
	jest.clearAllMocks();
	mockLocation.pathname = '/videos';
	mockLocation.search = '';
	playlistsData = [playlist, folderPlaylist, clipPlaylist, moviePlaylist, personalPlaylist];
	allVideosData = [{ id: 30, name: 'ep-1.mp4', parent_path: '/videos', format: 'mp4' }];
	homeCatalogData = {
		sections: [
			{
				key: 'recent',
				items: [{ video: { id: 40, name: 'recent.mp4', parent_path: '/recent', format: 'mp4' }, status: 'not_started', progress_pct: 0 }],
			},
		],
	};
	playbackData = { playback_state: { playlist_id: 1, video_id: 30 } };
	membershipData = {};
	selectedPlaylistData = detailPlaylist;
	selectedPlaylistLoading = false;
	mutationShouldError = false;

	mockUseQueryClient.mockReturnValue({ invalidateQueries: mockInvalidateQueries });
	mockGetVideoPlaylists.mockResolvedValue(playlistsData);
	mockGetAllVideoFiles.mockResolvedValue(allVideosData);
	mockGetVideoHomeCatalog.mockResolvedValue(homeCatalogData);
	mockGetVideoPlaylistById.mockResolvedValue(detailPlaylist);
	mockGetVideoPlaybackState.mockResolvedValue(playbackData);
	mockAddVideoToPlaylist.mockResolvedValue({});
	mockRemoveVideoFromPlaylist.mockResolvedValue({});
	mockReorderVideoPlaylist.mockResolvedValue({});
	mockUpdateVideoPlaylistName.mockResolvedValue({});

	mockUseQuery.mockImplementation((options: any) => {
		const [key] = options.queryKey;
		if (options.enabled !== false) {
			options.queryFn?.();
		}
		if (key === 'video-playlists') return { data: playlistsData, isLoading: false };
		if (key === 'video-all-files') return { data: allVideosData, isLoading: false };
		if (key === 'video-home-catalog') return { data: homeCatalogData, isLoading: false };
		if (key === 'video-playback-state') return { data: playbackData };
		if (key === 'video-playlist-membership') return { data: membershipData };
		if (key === 'video-playlist') return { data: selectedPlaylistData, isLoading: selectedPlaylistLoading };
		return { data: undefined, isLoading: false };
	});

	mockUseMutation.mockImplementation((options: any) => ({
		mutate: (...args: any[]) => {
			options.mutationFn?.(...args);
			if (mutationShouldError) {
				options.onError?.(...args);
				return;
			}
			options.onSuccess?.(...args);
		},
		isPending: false,
	}));
});

describe('components/videos/videoContent', () => {
	it('renders loading state', () => {
		mockUseQuery.mockImplementation((options: any) => {
			const [key] = options.queryKey;
			if (key === 'video-playlists') return { data: [], isLoading: true };
			if (key === 'video-all-files') return { data: [], isLoading: false };
			if (key === 'video-home-catalog') return { data: homeCatalogData, isLoading: false };
			if (key === 'video-playback-state') return { data: playbackData };
			if (key === 'video-playlist-membership') return { data: {} };
			return { data: undefined, isLoading: false };
		});

		render(<VideoContent />);
		expect(screen.getByText('Carregando videos...')).toBeInTheDocument();
	});

	it('renders video home with contextual sections and recent catalog', () => {
		render(<VideoContent />);

		expect(screen.getByText('Continuar assistindo')).toBeInTheDocument();
		expect(screen.getByText('Series')).toBeInTheDocument();
		expect(screen.getByText('Clipes')).toBeInTheDocument();
		expect(screen.getByText('Adicionados recentemente')).toBeInTheDocument();
		expect(screen.getByText('recent.mp4')).toBeInTheDocument();
		expect(screen.getAllByText('Abrir secao').length).toBeGreaterThan(0);
	});

	it('routes home card selection to contextual detail urls', () => {
		render(<VideoContent />);

		fireEvent.click(screen.getAllByRole('button', { name: /Playlist One/i })[0]!);
		fireEvent.click(screen.getAllByRole('button', { name: /Movie Playlist/i })[0]!);
		fireEvent.click(screen.getAllByRole('button', { name: /Personal Playlist/i })[0]!);
		fireEvent.click(screen.getAllByRole('button', { name: /Clip Playlist/i })[0]!);
		fireEvent.click(screen.getAllByRole('button', { name: /Folder Playlist/i })[0]!);

		expect(mockNavigate).toHaveBeenNthCalledWith(1, '/videos/series/playlist-one');
		expect(mockNavigate).toHaveBeenNthCalledWith(2, '/videos/movies/movie-playlist');
		expect(mockNavigate).toHaveBeenNthCalledWith(3, '/videos/personal/personal-playlist');
		expect(mockNavigate).toHaveBeenNthCalledWith(4, '/videos/clips/clip-playlist');
		expect(mockNavigate).toHaveBeenNthCalledWith(5, '/videos/folders/folder-playlist');
	});

	it('renders folders section and handles add/play/search actions', async () => {
		mockLocation.pathname = '/videos/folders';
		render(<VideoContent />);

		expect(screen.getByText('Pastas')).toBeInTheDocument();
		expect(screen.getAllByText('Folder Playlist').length).toBeGreaterThan(0);
		expect(screen.getByText('Todos')).toBeInTheDocument();

		fireEvent.change(screen.getByPlaceholderText('Buscar video por nome, pasta ou formato'), { target: { value: 'no-match' } });
		expect(screen.queryByText('ep-1.mp4')).not.toBeInTheDocument();
		fireEvent.change(screen.getByPlaceholderText('Buscar video por nome, pasta ou formato'), { target: { value: 'ep-1' } });
		expect(screen.getByText('ep-1.mp4')).toBeInTheDocument();

		fireEvent.click(screen.getAllByRole('button', { name: /Reproduzir/i })[0]!);
		expect(mockNavigate).toHaveBeenCalled();

		fireEvent.change(screen.getByRole('combobox'), { target: { value: '1' } });
		fireEvent.click(screen.getByRole('button', { name: /Adicionar/i }));
		expect(mockAddVideoToPlaylist).toHaveBeenCalledWith(1, 30);
		expect(await screen.findByText('Video adicionado a playlist com sucesso.')).toBeInTheDocument();
	});

	it('renders series detail branch from route-based detail path', () => {
		mockLocation.pathname = '/videos/series/playlist-one';

		render(<VideoContent />);
		expect(screen.getByText('Temporada 1')).toBeInTheDocument();
		expect(screen.getAllByText('Show').length).toBeGreaterThan(0);
		expect(screen.getByText('Retomar agora')).toBeInTheDocument();

		fireEvent.click(screen.getByRole('button', { name: /Retomar agora/i }));
		expect(mockNavigate).toHaveBeenCalledWith('/video/31', {
			state: { from: '/videos/series/playlist-one', playlistId: 1 },
		});
	});

	it('renders folder detail branch and actions', () => {
		mockLocation.pathname = '/videos/folders/folder-playlist';
		selectedPlaylistData = folderDetailPlaylist;

		render(<VideoContent />);
		expect(screen.getByText('Editar playlist')).toBeInTheDocument();

		fireEvent.change(screen.getByPlaceholderText('Nome de exibicao'), { target: { value: 'Renamed' } });
		fireEvent.click(screen.getByRole('button', { name: 'Salvar nome' }));
		expect(mockUpdateVideoPlaylistName).toHaveBeenCalledWith(2, 'Renamed');

		fireEvent.click(screen.getByRole('button', { name: /Voltar para videos/i }));
		expect(mockNavigate).toHaveBeenCalledWith('/videos/folders');
	});

	it('renders context detail branch for clips', () => {
		mockLocation.pathname = '/videos/clips/clip-playlist';
		selectedPlaylistData = clipDetailPlaylist;

		render(<VideoContent />);
		expect(screen.getByText('Itens da colecao')).toBeInTheDocument();
		expect(screen.getByText('Quick Clip')).toBeInTheDocument();

		fireEvent.click(screen.getByRole('button', { name: /Reproduzir/i }));
		expect(mockNavigate).toHaveBeenCalledWith('/video/32', {
			state: { from: '/videos/clips/clip-playlist', playlistId: 3 },
		});
	});

	it('handles add-to-playlist error', async () => {
		mockLocation.pathname = '/videos/folders';
		mutationShouldError = true;
		render(<VideoContent />);

		fireEvent.click(screen.getByRole('button', { name: /Adicionar/i }));
		expect(await screen.findByText('Nao foi possivel adicionar o video a playlist.')).toBeInTheDocument();
	});
});
