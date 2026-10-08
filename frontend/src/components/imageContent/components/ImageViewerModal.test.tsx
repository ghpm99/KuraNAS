import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen } from '@testing-library/react';
import type { ReactElement } from 'react';
import type { ImageLibraryItem } from '@/types/imageLibrary';
import { buildImageLibraryItem } from '../imageLibraryTestFixtures';
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
                IMAGES_VIEWER_DOWNLOAD: 'Baixar original',
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

const createImage = (overrides: Partial<ImageLibraryItem> = {}) => buildImageLibraryItem(overrides);

const renderWithQuery = (ui: ReactElement) =>
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            {ui}
        </QueryClientProvider>
    );

describe('ImageViewerModal', () => {
    it('renders product actions, details, and filmstrip items', () => {
        const onToggleFavorite = jest.fn();
        const onOpenFolder = jest.fn();

        renderWithQuery(
            <ImageViewerModal
                activeImage={createImage()}
                activeIndex={0}
                activeImageDate={new Date('2026-03-10T10:00:00Z')}
                dateFormatter={
                    new Intl.DateTimeFormat('pt-BR', {
                        dateStyle: 'medium',
                        timeStyle: 'short',
                    })
                }
                filteredImages={[createImage(), createImage({ file_id: 8, name: 'Trip-2.jpg' })]}
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
            />
        );

        expect(screen.getByRole('button', { name: 'Favoritar' })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Abrir pasta' })).toBeInTheDocument();
        expect(screen.getByText('Biblioteca')).toBeInTheDocument();
        expect(screen.getAllByText('/photos/travel')).toHaveLength(2);
        expect(screen.getByText('Foto')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Abrir Trip-2\.jpg/i })).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'Favoritar' }));
        fireEvent.click(screen.getByRole('button', { name: 'Abrir pasta' }));

        expect(onToggleFavorite).toHaveBeenCalledTimes(1);
        expect(onOpenFolder).toHaveBeenCalledTimes(1);
    });

    it('shows the preview at 1x and swaps to the original when zoomed in', () => {
        const baseProps = {
            activeImage: createImage({ file_id: 11 }),
            activeIndex: 0,
            activeImageDate: null,
            dateFormatter: new Intl.DateTimeFormat('pt-BR'),
            filteredImages: [createImage({ file_id: 11 })],
            showDetails: false,
            showFilmstrip: true,
            isSlideshowPlaying: false,
            isFavoritePending: false,
            onToggleDetails: jest.fn(),
            onToggleFilmstrip: jest.fn(),
            onToggleSlideshow: jest.fn(),
            onToggleFavorite: jest.fn(),
            onOpenFolder: jest.fn(),
            onDecreaseZoom: jest.fn(),
            onResetZoom: jest.fn(),
            onIncreaseZoom: jest.fn(),
            onClose: jest.fn(),
            onPrevious: jest.fn(),
            onNext: jest.fn(),
            onOpenImage: jest.fn(),
        };

        const { rerender } = renderWithQuery(<ImageViewerModal {...baseProps} zoom={1} />);
        const stageImage = () =>
            document.querySelector<HTMLImageElement>('img[class*="image"]') as HTMLImageElement;

        expect(stageImage().getAttribute('src')).toBe(
            '/api/v1/files/thumbnail/11?width=1600&height=1600'
        );
        expect(document.querySelector('button img')?.getAttribute('src')).toBe(
            '/api/v1/files/thumbnail/11?width=160&height=160'
        );

        rerender(
            <QueryClientProvider client={new QueryClient()}>
                <ImageViewerModal {...baseProps} zoom={1.4} />
            </QueryClientProvider>
        );

        expect(stageImage().getAttribute('src')).toBe('/api/v1/files/blob/11');
    });

    it.each([
        ['.heic', false],
        ['.HEIF', false],
        ['.tiff', false],
        ['.tif', false],
        ['.cr2', false],
        ['.dng', false],
        ['.raf', false],
        ['.jpg', true],
        ['.jfif', true],
        ['.avif', true],
        ['.webp', true],
    ])('format %s swaps to the original on zoom: %s', (format, canSwapToOriginal) => {
        const activeImage = createImage({ file_id: 21, format });
        renderWithQuery(
            <ImageViewerModal
                activeImage={activeImage}
                activeIndex={0}
                activeImageDate={null}
                dateFormatter={new Intl.DateTimeFormat('pt-BR')}
                filteredImages={[activeImage]}
                zoom={2}
                showDetails={false}
                showFilmstrip={false}
                isSlideshowPlaying={false}
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
            />
        );

        const stageImage = document.querySelector<HTMLImageElement>('img[class*="image"]');
        expect(stageImage?.getAttribute('src')).toBe(
            canSwapToOriginal
                ? '/api/v1/files/blob/21'
                : '/api/v1/files/thumbnail/21?width=1600&height=1600'
        );
        expect(screen.getByRole('link', { name: 'Baixar original' }).getAttribute('href')).toBe(
            '/api/v1/files/download/21'
        );
    });

    it('shows the playing state and hides the filmstrip when requested', () => {
        renderWithQuery(
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
            />
        );

        expect(screen.getByRole('button', { name: 'Desfavoritar' })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Pausar slideshow' })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Mostrar tira' })).toBeInTheDocument();
        expect(screen.queryByRole('button', { name: /Abrir Trip\.jpg/i })).not.toBeInTheDocument();
    });

    it('calls onIncreaseZoom on wheel up (deltaY < 0) and onDecreaseZoom on wheel down (deltaY > 0)', () => {
        const onIncreaseZoom = jest.fn();
        const onDecreaseZoom = jest.fn();

        renderWithQuery(
            <ImageViewerModal
                activeImage={createImage()}
                activeIndex={0}
                activeImageDate={new Date('2026-03-10T10:00:00Z')}
                dateFormatter={new Intl.DateTimeFormat('pt-BR')}
                filteredImages={[createImage()]}
                zoom={1}
                showDetails={false}
                showFilmstrip={false}
                isSlideshowPlaying={false}
                isFavoritePending={false}
                onToggleDetails={jest.fn()}
                onToggleFilmstrip={jest.fn()}
                onToggleSlideshow={jest.fn()}
                onToggleFavorite={jest.fn()}
                onOpenFolder={jest.fn()}
                onDecreaseZoom={onDecreaseZoom}
                onResetZoom={jest.fn()}
                onIncreaseZoom={onIncreaseZoom}
                onClose={jest.fn()}
                onPrevious={jest.fn()}
                onNext={jest.fn()}
                onOpenImage={jest.fn()}
            />
        );

        const stagePanel = screen.getByRole('button', {
            name: 'Imagem anterior',
        }).parentElement!;

        fireEvent.wheel(stagePanel, { deltaY: -100 });
        expect(onIncreaseZoom).toHaveBeenCalledTimes(1);
        expect(onDecreaseZoom).not.toHaveBeenCalled();

        onIncreaseZoom.mockClear();
        fireEvent.wheel(stagePanel, { deltaY: 100 });
        expect(onDecreaseZoom).toHaveBeenCalledTimes(1);
        expect(onIncreaseZoom).not.toHaveBeenCalled();
    });

    it('renders filmstrip with non-active items and triggers onOpenImage', () => {
        const onOpenImage = jest.fn();
        const activeImage = createImage({ file_id: 7 });
        const otherImage = createImage({ file_id: 8, name: 'Other.jpg' });

        renderWithQuery(
            <ImageViewerModal
                activeImage={activeImage}
                activeIndex={0}
                activeImageDate={new Date('2026-03-10T10:00:00Z')}
                dateFormatter={new Intl.DateTimeFormat('pt-BR')}
                filteredImages={[activeImage, otherImage]}
                zoom={1}
                showDetails={false}
                showFilmstrip
                isSlideshowPlaying={false}
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
                onOpenImage={onOpenImage}
            />
        );

        const otherButton = screen.getByRole('button', {
            name: /Abrir Other\.jpg/i,
        });
        expect(otherButton).toBeInTheDocument();

        fireEvent.click(otherButton);
        expect(onOpenImage).toHaveBeenCalledWith(8);
    });

    it('disables slideshow button when only one image', () => {
        renderWithQuery(
            <ImageViewerModal
                activeImage={createImage()}
                activeIndex={0}
                activeImageDate={null}
                dateFormatter={new Intl.DateTimeFormat('pt-BR')}
                filteredImages={[createImage()]}
                zoom={1}
                showDetails={false}
                showFilmstrip={false}
                isSlideshowPlaying={false}
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
            />
        );

        const slideshowButton = screen.getByRole('button', {
            name: 'Iniciar slideshow',
        });
        expect(slideshowButton).toBeDisabled();
    });
});
