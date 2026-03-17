import { renderHook } from '@testing-library/react';
import type { IImageData } from '@/components/providers/imageProvider/imageProvider';
import { useImageViewerModal } from './useImageViewerModal';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string, params?: Record<string, string>) => {
			const map: Record<string, string> = {
				COMMON_NOT_AVAILABLE: 'N/D',
				IMAGES_DATE_UNAVAILABLE: 'Data indisponivel',
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
		classification: { category: 'photo' as const, confidence: 0.95 },
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

const dateFormatter = new Intl.DateTimeFormat('pt-BR', { dateStyle: 'medium', timeStyle: 'short' });

describe('useImageViewerModal', () => {
	it('returns detail sections with full metadata', () => {
		const activeImage = createImage();
		const activeImageDate = new Date('2026-03-10T10:00:00Z');

		const { result } = renderHook(() =>
			useImageViewerModal({
				activeImage,
				activeImageDate,
				activeIndex: 2,
				totalImages: 10,
				dateFormatter,
			}),
		);

		expect(result.current.folderPath).toBe('/photos/travel');
		expect(result.current.positionLabel).toBe('3 de 10');

		const [librarySection, captureSection, deviceSection] = result.current.details;

		// Library section
		expect(librarySection.title).toBe('Biblioteca');
		const folderItem = librarySection.items.find((i) => i.label === 'Pasta');
		expect(folderItem?.value).toBe('/photos/travel');
		const formatItem = librarySection.items.find((i) => i.label === 'Formato');
		expect(formatItem?.value).toBe('.jpg');
		const dimensionsItem = librarySection.items.find((i) => i.label === 'Dimensoes');
		expect(dimensionsItem?.value).toBe('1600 x 900');
		const categoryItem = librarySection.items.find((i) => i.label === 'Categoria');
		expect(categoryItem?.value).toBe('Foto');
		const confidenceItem = librarySection.items.find((i) => i.label === 'Confianca');
		expect(confidenceItem?.value).toBe('95%');

		// Capture section
		expect(captureSection.title).toBe('Captura');
		const dateItem = captureSection.items.find((i) => i.label === 'Data');
		expect(dateItem?.value).not.toBe('Data indisponivel');
		const softwareItem = captureSection.items.find((i) => i.label === 'Software');
		expect(softwareItem?.value).toBe('Photos App');
		const descriptionItem = captureSection.items.find((i) => i.label === 'Descricao');
		expect(descriptionItem?.value).toBe('Trip');

		// Device section
		expect(deviceSection.title).toBe('Dispositivo');
		const cameraItem = deviceSection.items.find((i) => i.label === 'Camera');
		expect(cameraItem?.value).toBe('Sony A7');
		const lensItem = deviceSection.items.find((i) => i.label === 'Lente');
		expect(lensItem?.value).toBe('24-70mm');
		const isoItem = deviceSection.items.find((i) => i.label === 'ISO');
		expect(isoItem?.value).toBe('400');
		const focalItem = deviceSection.items.find((i) => i.label === 'Focal');
		expect(focalItem?.value).toBe('35mm');
		const apertureItem = deviceSection.items.find((i) => i.label === 'Abertura');
		expect(apertureItem?.value).toBe('f/2.8');
		const exposureItem = deviceSection.items.find((i) => i.label === 'Exposicao');
		expect(exposureItem?.value).toBe('1/125s');
	});

	it('falls back to N/D when metadata is missing', () => {
		const activeImage = createImage({
			format: '',
			metadata: {
				width: 0,
				height: 0,
				make: '',
				model: '',
				lens_model: '',
				software: '',
				image_description: '',
				iso: 0,
				focal_length: 0,
				f_number: 0,
				exposure_time: 0,
				classification: undefined as any,
			},
		});

		const { result } = renderHook(() =>
			useImageViewerModal({
				activeImage,
				activeImageDate: null,
				activeIndex: 0,
				totalImages: 1,
				dateFormatter,
			}),
		);

		const [librarySection, captureSection, deviceSection] = result.current.details;

		const dimensionsItem = librarySection.items.find((i) => i.label === 'Dimensoes');
		expect(dimensionsItem?.value).toBe('N/D');

		const dateItem = captureSection.items.find((i) => i.label === 'Data');
		expect(dateItem?.value).toBe('Data indisponivel');

		const cameraItem = deviceSection.items.find((i) => i.label === 'Camera');
		expect(cameraItem?.value).toBe('N/D');
		const lensItem = deviceSection.items.find((i) => i.label === 'Lente');
		expect(lensItem?.value).toBe('N/D');
		const isoItem = deviceSection.items.find((i) => i.label === 'ISO');
		expect(isoItem?.value).toBe('N/D');
		const focalItem = deviceSection.items.find((i) => i.label === 'Focal');
		expect(focalItem?.value).toBe('N/D');
		const apertureItem = deviceSection.items.find((i) => i.label === 'Abertura');
		expect(apertureItem?.value).toBe('N/D');
		const exposureItem = deviceSection.items.find((i) => i.label === 'Exposicao');
		expect(exposureItem?.value).toBe('N/D');
	});

	it('handles image with no metadata at all', () => {
		const activeImage = createImage({ metadata: undefined as any });
		// Remove metadata entirely
		(activeImage as any).metadata = undefined;

		const { result } = renderHook(() =>
			useImageViewerModal({
				activeImage,
				activeImageDate: null,
				activeIndex: 0,
				totalImages: 5,
				dateFormatter,
			}),
		);

		const [librarySection, , deviceSection] = result.current.details;

		const dimensionsItem = librarySection.items.find((i) => i.label === 'Dimensoes');
		expect(dimensionsItem?.value).toBe('N/D');
		const categoryItem = librarySection.items.find((i) => i.label === 'Categoria');
		expect(categoryItem?.value).toBe('Outros');
		const confidenceItem = librarySection.items.find((i) => i.label === 'Confianca');
		expect(confidenceItem?.value).toBe('N/D');

		const cameraItem = deviceSection.items.find((i) => i.label === 'Camera');
		expect(cameraItem?.value).toBe('N/D');
	});

	it('uses "capture" classification label key', () => {
		const activeImage = createImage({
			metadata: {
				classification: { category: 'capture', confidence: 0.88 },
			},
		});

		const { result } = renderHook(() =>
			useImageViewerModal({
				activeImage,
				activeImageDate: new Date('2026-03-10T10:00:00Z'),
				activeIndex: 0,
				totalImages: 1,
				dateFormatter,
			}),
		);

		const [librarySection] = result.current.details;
		const categoryItem = librarySection.items.find((i) => i.label === 'Categoria');
		expect(categoryItem?.value).toBe('Captura');
		const confidenceItem = librarySection.items.find((i) => i.label === 'Confianca');
		expect(confidenceItem?.value).toBe('88%');
	});

	it('uses "other" classification label key', () => {
		const activeImage = createImage({
			metadata: {
				classification: { category: 'other', confidence: 0.5 },
			},
		});

		const { result } = renderHook(() =>
			useImageViewerModal({
				activeImage,
				activeImageDate: null,
				activeIndex: 0,
				totalImages: 1,
				dateFormatter,
			}),
		);

		const [librarySection] = result.current.details;
		const categoryItem = librarySection.items.find((i) => i.label === 'Categoria');
		expect(categoryItem?.value).toBe('Outros');
	});

	it('formats exposure time >= 1 as plain seconds', () => {
		const activeImage = createImage({
			metadata: { exposure_time: 2 },
		});

		const { result } = renderHook(() =>
			useImageViewerModal({
				activeImage,
				activeImageDate: null,
				activeIndex: 0,
				totalImages: 1,
				dateFormatter,
			}),
		);

		const [, , deviceSection] = result.current.details;
		const exposureItem = deviceSection.items.find((i) => i.label === 'Exposicao');
		expect(exposureItem?.value).toBe('2s');
	});

	it('formats exposure time exactly equal to 1', () => {
		const activeImage = createImage({
			metadata: { exposure_time: 1 },
		});

		const { result } = renderHook(() =>
			useImageViewerModal({
				activeImage,
				activeImageDate: null,
				activeIndex: 0,
				totalImages: 1,
				dateFormatter,
			}),
		);

		const [, , deviceSection] = result.current.details;
		const exposureItem = deviceSection.items.find((i) => i.label === 'Exposicao');
		expect(exposureItem?.value).toBe('1s');
	});

	it('formats non-integer focal length and f-number with one decimal', () => {
		const activeImage = createImage({
			metadata: { focal_length: 50.5, f_number: 1.4 },
		});

		const { result } = renderHook(() =>
			useImageViewerModal({
				activeImage,
				activeImageDate: null,
				activeIndex: 0,
				totalImages: 1,
				dateFormatter,
			}),
		);

		const [, , deviceSection] = result.current.details;
		const focalItem = deviceSection.items.find((i) => i.label === 'Focal');
		expect(focalItem?.value).toBe('50.5mm');
		const apertureItem = deviceSection.items.find((i) => i.label === 'Abertura');
		expect(apertureItem?.value).toBe('f/1.4');
	});

	it('shows only make when model is empty', () => {
		const activeImage = createImage({
			metadata: { make: 'Canon', model: '' },
		});

		const { result } = renderHook(() =>
			useImageViewerModal({
				activeImage,
				activeImageDate: null,
				activeIndex: 0,
				totalImages: 1,
				dateFormatter,
			}),
		);

		const [, , deviceSection] = result.current.details;
		const cameraItem = deviceSection.items.find((i) => i.label === 'Camera');
		expect(cameraItem?.value).toBe('Canon');
	});

	it('shows only model when make is empty', () => {
		const activeImage = createImage({
			metadata: { make: '', model: 'A7' },
		});

		const { result } = renderHook(() =>
			useImageViewerModal({
				activeImage,
				activeImageDate: null,
				activeIndex: 0,
				totalImages: 1,
				dateFormatter,
			}),
		);

		const [, , deviceSection] = result.current.details;
		const cameraItem = deviceSection.items.find((i) => i.label === 'Camera');
		expect(cameraItem?.value).toBe('A7');
	});

	it('shows N/D for created_at when missing', () => {
		const activeImage = createImage({ created_at: '' });

		const { result } = renderHook(() =>
			useImageViewerModal({
				activeImage,
				activeImageDate: new Date('2026-03-10T10:00:00Z'),
				activeIndex: 0,
				totalImages: 1,
				dateFormatter,
			}),
		);

		const [, captureSection] = result.current.details;
		const createdItem = captureSection.items.find((i) => i.label === 'Criado em');
		expect(createdItem?.value).toBe('N/D');
	});

	it('shows format value when present but trims whitespace', () => {
		const activeImage = createImage({ format: '  .png  ' });

		const { result } = renderHook(() =>
			useImageViewerModal({
				activeImage,
				activeImageDate: null,
				activeIndex: 0,
				totalImages: 1,
				dateFormatter,
			}),
		);

		const [librarySection] = result.current.details;
		const formatItem = librarySection.items.find((i) => i.label === 'Formato');
		expect(formatItem?.value).toBe('.png');
	});

	it('handles width present but height missing for resolution', () => {
		const activeImage = createImage({
			metadata: { width: 1920, height: 0 },
		});

		const { result } = renderHook(() =>
			useImageViewerModal({
				activeImage,
				activeImageDate: null,
				activeIndex: 0,
				totalImages: 1,
				dateFormatter,
			}),
		);

		const [librarySection] = result.current.details;
		const dimensionsItem = librarySection.items.find((i) => i.label === 'Dimensoes');
		expect(dimensionsItem?.value).toBe('N/D');
	});

	it('handles height present but width missing for resolution', () => {
		const activeImage = createImage({
			metadata: { width: 0, height: 1080 },
		});

		const { result } = renderHook(() =>
			useImageViewerModal({
				activeImage,
				activeImageDate: null,
				activeIndex: 0,
				totalImages: 1,
				dateFormatter,
			}),
		);

		const [librarySection] = result.current.details;
		const dimensionsItem = librarySection.items.find((i) => i.label === 'Dimensoes');
		expect(dimensionsItem?.value).toBe('N/D');
	});

	it('handles image path with no directory structure', () => {
		const activeImage = createImage({ path: 'image.jpg', name: 'image.jpg' });

		const { result } = renderHook(() =>
			useImageViewerModal({
				activeImage,
				activeImageDate: null,
				activeIndex: 0,
				totalImages: 1,
				dateFormatter,
			}),
		);

		expect(result.current.folderPath).toBe('/');
	});
});
