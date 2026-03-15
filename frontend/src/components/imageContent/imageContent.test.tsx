import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import ImageContent from './imageContent';

const mockUseImage = jest.fn();
const mockRef = jest.fn();
const mockUseIntersectionObserver = jest.fn();
const mockToggleStarredFile = jest.fn();
const mockNavigate = jest.fn();
const mockEnqueueSnackbar = jest.fn();

jest.mock('../providers/imageProvider/imageProvider', () => ({ useImage: () => mockUseImage() }));
jest.mock('../hooks/IntersectionObserver/useIntersectionObserver', () => ({
	useIntersectionObserver: (...args: any[]) => mockUseIntersectionObserver(...args),
}));
jest.mock('notistack', () => ({
	useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));
jest.mock('@/service/files', () => ({
	toggleStarredFile: (...args: any[]) => mockToggleStarredFile(...args),
}));
jest.mock('react-router-dom', () => {
	const actual = jest.requireActual('react-router-dom');
	return {
		...actual,
		useNavigate: () => mockNavigate,
	};
});
jest.mock('@/service/apiUrl', () => ({
	getApiV1BaseUrl: () => '/api/v1',
}));
jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string, params?: Record<string, string | number>) => {
			const map: Record<string, string> = {
				LOCALE: 'pt-BR',
				IMAGES_SECTION_LIBRARY: 'Biblioteca',
				IMAGES_SECTION_CAPTURES: 'Capturas',
				IMAGES_SECTION_FOLDERS: 'Pastas',
				IMAGES_SECTION_ALBUMS: 'Albuns automaticos',
				IMAGES_ALBUM_OTHERS: 'Outros',
				IMAGES_ALBUM_OTHERS_DESCRIPTION: 'Tudo que nao entrou nos temas principais',
				IMAGES_FOLDERS_SUMMARY: `${params?.filtered ?? 0} de ${params?.total ?? 0} pastas`,
				IMAGES_ALBUMS_SUMMARY: `${params?.filtered ?? 0} de ${params?.total ?? 0} albuns`,
				IMAGES_COUNT_SUMMARY: `${params?.filtered ?? 0} de ${params?.total ?? 0} imagens`,
				IMAGES_END_MESSAGE: 'Todas as imagens carregadas',
				IMAGES_COLLECTION_OPEN: `Abrir ${params?.name ?? ''}`.trim(),
				IMAGES_OPEN_IMAGE_ARIA: `Abrir ${params?.name ?? ''}`.trim(),
				IMAGES_DETAILS_TITLE: 'Detalhes',
				IMAGES_CLOSE_VIEWER: 'Fechar visualizador',
				IMAGES_GROUP_BY_ARIA: 'Agrupar imagens por',
				IMAGES_BACK_TO_FOLDERS: 'Voltar para pastas',
				IMAGES_BACK_TO_ALBUMS: 'Voltar para albuns',
				IMAGES_VIEWER_ADD_FAVORITE: 'Favoritar',
				IMAGES_VIEWER_REMOVE_FAVORITE: 'Desfavoritar',
				IMAGES_VIEWER_OPEN_FOLDER: 'Abrir pasta',
				IMAGES_VIEWER_START_SLIDESHOW: 'Iniciar slideshow',
				IMAGES_VIEWER_STOP_SLIDESHOW: 'Pausar slideshow',
				IMAGES_VIEWER_HIDE_FILMSTRIP: 'Ocultar tira',
				IMAGES_VIEWER_SHOW_FILMSTRIP: 'Mostrar tira',
				IMAGES_VIEWER_HIDE_FILMSTRIP_SHORT: 'Tira off',
				IMAGES_VIEWER_SHOW_FILMSTRIP_SHORT: 'Tira on',
				IMAGES_VIEWER_KEYBOARD_HINT: 'Atalhos',
				IMAGES_VIEWER_POSITION: `${params?.current ?? 1} de ${params?.total ?? 1}`,
				IMAGES_VIEWER_FAVORITE_ADDED: 'Imagem adicionada aos favoritos',
				IMAGES_VIEWER_FAVORITE_REMOVED: 'Imagem removida dos favoritos',
				IMAGES_VIEWER_FAVORITE_ERROR: 'Erro ao atualizar favorito',
				IMAGES_DETAILS_SECTION_LIBRARY: 'Biblioteca',
				IMAGES_DETAILS_SECTION_CAPTURE: 'Captura',
				IMAGES_DETAILS_SECTION_DEVICE: 'Dispositivo',
				IMAGES_DETAIL_DATE: 'Data',
				IMAGES_DETAIL_CREATED: 'Criado em',
				IMAGES_DETAIL_SOFTWARE: 'Software',
				IMAGES_DETAIL_DESCRIPTION: 'Descricao',
				IMAGES_DETAIL_CATEGORY: 'Categoria',
				IMAGES_DETAIL_CONFIDENCE: 'Confianca',
				IMAGES_CLASSIFICATION_CAPTURE: 'Captura',
				IMAGES_CLASSIFICATION_PHOTO: 'Foto',
				IMAGES_CLASSIFICATION_OTHER: 'Outro',
				IMAGES_FOLDERS_EMPTY_TITLE: 'Nenhuma pasta encontrada',
				IMAGES_FOLDERS_EMPTY_DESC: 'Sem pastas',
				IMAGES_ALBUMS_EMPTY_TITLE: 'Nenhum album encontrado',
				IMAGES_ALBUMS_EMPTY_DESC: 'Sem albuns',
			};
			return map[key] ?? key;
		},
	}),
}));
jest.mock('@/components/providers/settingsProvider/settingsContext', () => ({
	useSettings: () => ({
		settings: {
			library: { runtime_root_path: '/data', watched_paths: ['/data'], remember_last_location: true, prioritize_favorites: true },
			indexing: { workers_enabled: true, scan_on_startup: true, extract_metadata: true, generate_previews: true },
			players: { remember_music_queue: true, remember_video_progress: true, autoplay_next_video: true, image_slideshow_seconds: 4 },
			appearance: { accent_color: 'violet', reduce_motion: false },
			language: { current: 'pt-BR', available: ['en-US', 'pt-BR'] },
		},
	}),
}));

const createQueryClient = () =>
	new QueryClient({
		defaultOptions: {
			queries: {
				retry: false,
			},
		},
	});

const renderImageContent = (initialEntries: string[]) =>
	render(
		<QueryClientProvider client={createQueryClient()}>
			<MemoryRouter initialEntries={initialEntries}>
				<ImageContent />
			</MemoryRouter>
		</QueryClientProvider>,
	);

const createImage = (overrides: Record<string, any> = {}) => ({
	id: 1,
	name: 'img1',
	path: '/photos/img1.jpg',
	format: '.jpg',
	size: 1024,
	created_at: '2026-03-10T10:00:00Z',
	updated_at: '2026-03-10T10:00:00Z',
	metadata: {
		width: 1600,
		height: 900,
		make: 'Sony',
		model: 'A7',
		classification: { category: 'photo', confidence: 0.9 },
		...overrides.metadata,
	},
	...overrides,
});

describe('imageContent', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockUseIntersectionObserver.mockImplementation(() => ({ ref: mockRef }));
		mockToggleStarredFile.mockResolvedValue(undefined);
		mockUseImage.mockReturnValue({
			images: [createImage()],
			status: 'success',
			imageGroupBy: 'date',
			setImageGroupBy: jest.fn(),
			fetchNextPage: jest.fn(),
			hasNextPage: false,
			isFetchingNextPage: true,
		});
	});

	it('renders grouped library and opens viewer with details', () => {
		renderImageContent(['/images']);

		expect(screen.getByText('Biblioteca')).toBeInTheDocument();
		expect(screen.getByText('Todas as imagens carregadas')).toBeInTheDocument();
		expect(screen.getByRole('progressbar')).toBeInTheDocument();

		fireEvent.click(screen.getByRole('button', { name: /abrir img1/i }));
		expect(screen.getByRole('dialog')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Abrir pasta' })).toBeInTheDocument();

		fireEvent.click(screen.getByRole('button', { name: 'Fechar visualizador' }));
		expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
	});

	it('triggers infinite load on intersect when enabled', () => {
		const fetchNextPage = jest.fn();
		let optionsRef: any;
		mockUseIntersectionObserver.mockImplementation((options: any) => {
			optionsRef = options;
			return { ref: mockRef };
		});

		mockUseImage.mockReturnValue({
			images: [createImage()],
			status: 'success',
			imageGroupBy: 'date',
			setImageGroupBy: jest.fn(),
			fetchNextPage,
			hasNextPage: true,
			isFetchingNextPage: false,
		});

		renderImageContent(['/images']);
		optionsRef.onIntersect();

		expect(fetchNextPage).toHaveBeenCalled();
		expect(screen.queryByText('Todas as imagens carregadas')).not.toBeInTheDocument();
	});

	it('changes grouping through selector', () => {
		const setImageGroupBy = jest.fn();
		mockUseImage.mockReturnValue({
			images: [createImage()],
			status: 'success',
			imageGroupBy: 'date',
			setImageGroupBy,
			fetchNextPage: jest.fn(),
			hasNextPage: false,
			isFetchingNextPage: false,
		});

		renderImageContent(['/images?image=1']);
		fireEvent.change(screen.getByLabelText('Agrupar imagens por'), { target: { value: 'type' } });

		expect(setImageGroupBy).toHaveBeenCalledWith('type');
	});

	it('uses persisted backend classification for the captures route', () => {
		mockUseImage.mockReturnValue({
			images: [
				createImage({
					id: 1,
					name: 'Screenshot_local.png',
					path: '/photos/Screenshot_local.png',
					format: '.png',
					metadata: {
						width: 1600,
						height: 900,
						classification: { category: 'other', confidence: 0.2 },
					},
				}),
				createImage({
					id: 2,
					name: 'Trip.jpg',
					path: '/photos/Trip.jpg',
					metadata: {
						width: 1600,
						height: 900,
						classification: { category: 'capture', confidence: 0.98 },
					},
				}),
			],
			status: 'success',
			imageGroupBy: 'date',
			setImageGroupBy: jest.fn(),
			fetchNextPage: jest.fn(),
			hasNextPage: false,
			isFetchingNextPage: false,
		});

		renderImageContent(['/images/captures']);

		expect(screen.queryByRole('button', { name: /abrir screenshot_local\.png/i })).not.toBeInTheDocument();
		expect(screen.getByRole('button', { name: /abrir trip\.jpg/i })).toBeInTheDocument();
	});

	it('renders folder overview and allows entering a folder collection', () => {
		mockUseImage.mockReturnValue({
			images: [
				createImage({ id: 1, name: 'Trip.jpg', path: '/photos/travel/Trip.jpg' }),
				createImage({ id: 2, name: 'Family.jpg', path: '/photos/family/Family.jpg' }),
			],
			status: 'success',
			imageGroupBy: 'date',
			setImageGroupBy: jest.fn(),
			fetchNextPage: jest.fn(),
			hasNextPage: false,
			isFetchingNextPage: false,
		});

		renderImageContent(['/images/folders']);

		expect(screen.getByText('Pastas')).toBeInTheDocument();
		fireEvent.click(screen.getByRole('button', { name: /abrir travel/i }));

		expect(screen.getByText('/photos/travel')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Voltar para pastas' })).toBeInTheDocument();
	});

	it('renders album overview, enters a selected album, and returns to the overview', () => {
		mockUseImage.mockReturnValue({
			images: [
				createImage({
					id: 1,
					name: 'misc.jpg',
					path: '/photos/random/misc.jpg',
					metadata: {
						width: 1200,
						height: 900,
						classification: { category: 'photo', confidence: 0.9 },
					},
				}),
			],
			status: 'success',
			imageGroupBy: 'date',
			setImageGroupBy: jest.fn(),
			fetchNextPage: jest.fn(),
			hasNextPage: false,
			isFetchingNextPage: false,
		});

		renderImageContent(['/images/albums']);

		expect(screen.queryByLabelText('Agrupar imagens por')).not.toBeInTheDocument();
		fireEvent.click(screen.getByRole('button', { name: /abrir outros/i }));

		expect(screen.getByText('Tudo que nao entrou nos temas principais')).toBeInTheDocument();
		fireEvent.click(screen.getByRole('button', { name: 'Voltar para albuns' }));

		expect(screen.getByRole('button', { name: /abrir outros/i })).toBeInTheDocument();
	});

	it('renders the initial loading state for grid sections without images', () => {
		mockUseImage.mockReturnValue({
			images: [],
			status: 'pending',
			imageGroupBy: 'date',
			setImageGroupBy: jest.fn(),
			fetchNextPage: jest.fn(),
			hasNextPage: false,
			isFetchingNextPage: false,
		});

		renderImageContent(['/images']);

		expect(screen.getByRole('progressbar')).toBeInTheDocument();
		expect(screen.queryByText('IMAGES_EMPTY_TITLE')).not.toBeInTheDocument();
	});

	it('opens the viewer from the image search param', () => {
		renderImageContent(['/images?image=1']);

		expect(screen.getByRole('dialog')).toBeInTheDocument();
	});

	it('favorites the active image from the viewer and shows feedback', async () => {
		renderImageContent(['/images?image=1']);

		fireEvent.click(screen.getByRole('button', { name: 'Favoritar' }));

		await waitFor(() => expect(mockToggleStarredFile).toHaveBeenCalledWith(1));
	});

	it('opens the current folder in files from the viewer', () => {
		renderImageContent(['/images?image=1']);

		fireEvent.click(screen.getByRole('button', { name: 'Abrir pasta' }));

		expect(mockNavigate).toHaveBeenCalledWith({
			pathname: '/files',
			search: '?path=%2Fphotos',
		});
	});
});
