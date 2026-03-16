import { fireEvent, render, screen } from '@testing-library/react';
import type { IImageData } from '@/components/providers/imageProvider/imageProvider';
import ImageViewerModal from './ImageViewerModal';

jest.mock('@/service/apiUrl', () => ({
	getApiV1BaseUrl: () => '/api/v1',
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string, params?: Record<string, string>) => {
			const map: Record<string, string> = {
				COMMON_NOT_AVAILABLE: 'N/D',
				IMAGES_DATE_UNAVAILABLE: 'Data indisponivel',
				IMAGES_TOGGLE_DETAILS: 'Alternar detalhes',
				IMAGES_DECREASE_ZOOM: 'Reduzir zoom',
				IMAGES_RESET_ZOOM: 'Resetar zoom',
				IMAGES_INCREASE_ZOOM: 'Aumentar zoom',
				IMAGES_CLOSE_VIEWER: 'Fechar visualizador',
				IMAGES_PREVIOUS: 'Imagem anterior',
				IMAGES_NEXT: 'Proxima imagem',
				IMAGES_ZOOM_LABEL: 'Zoom',
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
				IMAGES_DETAILS_SECTION_LIBRARY: 'Biblioteca',
				IMAGES_DETAILS_SECTION_CAPTURE: 'Captura',
				IMAGES_DETAILS_SECTION_DEVICE: 'Dispositivo',
				IMAGES_DETAIL_FOLDER: 'Pasta',
				IMAGES_DETAIL_FORMAT: 'Formato',
				IMAGES_DETAIL_SIZE: 'Tamanho',
				IMAGES_DETAIL_DIMENSIONS: 'Dimensoes',
				IMAGES_DETAIL_CATEGORY: 'Categoria',
				IMAGES_DETAIL_CONFIDENCE: 'Confianca',
				IMAGES_DETAIL_DATE: 'Data',
				IMAGES_DETAIL_CREATED: 'Criado em',
				IMAGES_DETAIL_SOFTWARE: 'Software',
				IMAGES_DETAIL_DESCRIPTION: 'Descricao',
				IMAGES_DETAIL_CAMERA: 'Camera',
				IMAGES_DETAIL_LENS: 'Lente',
				IMAGES_DETAIL_ISO: 'ISO',
				IMAGES_DETAIL_FOCAL: 'Focal',
				IMAGES_DETAIL_APERTURE: 'Abertura',
				IMAGES_DETAIL_EXPOSURE: 'Exposicao',
				IMAGES_CLASSIFICATION_PHOTO: 'Foto',
				IMAGES_CLASSIFICATION_CAPTURE: 'Captura',
				IMAGES_CLASSIFICATION_OTHER: 'Outros',
				IMAGES_OPEN_IMAGE_ARIA: `Abrir ${params?.name ?? ''}`.trim(),
				IMAGES_VIEWER_POSITION: `${params?.current ?? '1'} de ${params?.total ?? '1'}`,
			};
			return map[key] ?? key;
		},
	}),
}));

const createImage = (
	overrides: Omit<Partial<IImageData>, 'metadata'> & { metadata?: Partial<NonNullable<IImageData['metadata']>> } = {},
): IImageData => {
	const metadata = {
		id: 7,
		fileId: 7,
		path: '/photos/travel/Trip.jpg',
		format: 'jpg',
		mode: 'RGB',
		width: 1600,
		height: 900,
		dpi_x: 72,
		dpi_y: 72,
		x_resolution: 72,
		y_resolution: 72,
		resolution_unit: 2,
		orientation: 1,
		compression: 0,
		photometric_interpretation: 0,
		color_space: 1,
		components_configuration: '',
		icc_profile: '',
		make: 'Sony',
		model: 'A7',
		lens_model: '24-70mm',
		serial_number: '',
		datetime: '2026-03-10T10:00:00Z',
		datetime_original: '2026-03-10T10:00:00Z',
		datetime_digitized: '',
		subsec_time: '',
		iso: 400,
		shutter_speed: 0,
		focal_length: 35,
		f_number: 2.8,
		aperture_value: 0,
		brightness_value: 0,
		exposure_bias: 0,
		metering_mode: 0,
		flash: 0,
		white_balance: 0,
		exposure_program: 0,
		max_aperture_value: 0,
		gps_latitude: 0,
		gps_longitude: 0,
		gps_altitude: 0,
		gps_date: '',
		gps_time: '',
		exposure_time: 0.008,
		user_comment: '',
		copyright: '',
		artist: '',
		software: 'Photos App',
		image_description: 'Trip',
		classification: { category: 'photo', confidence: 0.95 },
		createdAt: '2026-03-10T10:00:00Z',
		...overrides.metadata,
	} as NonNullable<IImageData['metadata']>;

	return {
		id: 7,
		name: 'Trip.jpg',
		path: '/photos/travel/Trip.jpg',
		type: 2,
		format: '.jpg',
		size: 2048,
		deleted_at: '',
		last_interaction: '',
		last_backup: '',
		check_sum: '',
		directory_content_count: 0,
		starred: false,
		created_at: '2026-03-10T10:00:00Z',
		updated_at: '2026-03-10T10:00:00Z',
		metadata,
		...overrides,
	} as IImageData;
};

describe('ImageViewerModal', () => {
	it('renders product actions, details, and filmstrip items', () => {
		const onToggleFavorite = jest.fn();
		const onOpenFolder = jest.fn();

		render(
			<ImageViewerModal
				activeImage={createImage()}
				activeIndex={0}
				activeImageDate={new Date('2026-03-10T10:00:00Z')}
				dateFormatter={new Intl.DateTimeFormat('pt-BR', { dateStyle: 'medium', timeStyle: 'short' })}
				filteredImages={[createImage(), createImage({ id: 8, name: 'Trip-2.jpg' })]}
				zoom={1}
				showDetails
				showFilmstrip
				isSlideshowPlaying={false}
				isFavoritePending={false}
				onToggleDetails={jest.fn()}
				onToggleFilmstrip={jest.fn()}
				onToggleSlideshow={jest.fn()}
				onToggleFavorite={onToggleFavorite}
				onOpenFolder={onOpenFolder}
				onDecreaseZoom={jest.fn()}
				onResetZoom={jest.fn()}
				onIncreaseZoom={jest.fn()}
				onClose={jest.fn()}
				onPrevious={jest.fn()}
				onNext={jest.fn()}
				onOpenImage={jest.fn()}
			/>,
		);

		expect(screen.getByRole('button', { name: 'Favoritar' })).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Abrir pasta' })).toBeInTheDocument();
		expect(screen.getByText('Biblioteca')).toBeInTheDocument();
		expect(screen.getAllByText('/photos/travel')).toHaveLength(2);
		expect(screen.getByRole('button', { name: /Abrir Trip-2\.jpg/i })).toBeInTheDocument();

		fireEvent.click(screen.getByRole('button', { name: 'Favoritar' }));
		fireEvent.click(screen.getByRole('button', { name: 'Abrir pasta' }));

		expect(onToggleFavorite).toHaveBeenCalledTimes(1);
		expect(onOpenFolder).toHaveBeenCalledTimes(1);
	});

	it('shows the playing state and hides the filmstrip when requested', () => {
		render(
			<ImageViewerModal
				activeImage={createImage({ starred: true })}
				activeIndex={0}
				activeImageDate={null}
				dateFormatter={new Intl.DateTimeFormat('pt-BR')}
				filteredImages={[createImage({ starred: true })]}
				zoom={1.4}
				showDetails={false}
				showFilmstrip={false}
				isSlideshowPlaying
				isFavoritePending={false}
				onToggleDetails={jest.fn()}
				onToggleFilmstrip={jest.fn()}
				onToggleSlideshow={jest.fn()}
				onToggleFavorite={jest.fn()}
				onOpenFolder={jest.fn()}
				onDecreaseZoom={jest.fn()}
				onResetZoom={jest.fn()}
				onIncreaseZoom={jest.fn()}
				onClose={jest.fn()}
				onPrevious={jest.fn()}
				onNext={jest.fn()}
				onOpenImage={jest.fn()}
			/>,
		);

		expect(screen.getByRole('button', { name: 'Desfavoritar' })).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Pausar slideshow' })).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Mostrar tira' })).toBeInTheDocument();
		expect(screen.queryByRole('button', { name: /Abrir Trip\.jpg/i })).not.toBeInTheDocument();
	});
});
