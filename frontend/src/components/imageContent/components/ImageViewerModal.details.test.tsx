import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { apiBase } from '@/service';
import { buildImageLibraryItem } from '../imageLibraryTestFixtures';
import ImageViewerModal from './ImageViewerModal';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn() },
}));

jest.mock('@/service/apiUrl', () => ({
    getApiV1BaseUrl: () => '/api/v1',
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string>) =>
            key === 'IMAGES_VIEWER_POSITION' ? `${params?.current}/${params?.total}` : key,
    }),
}));

const mockedApiGet = apiBase.get as jest.Mock;

const mockEndpoints = (summary: Record<string, unknown>, location: Record<string, unknown>) => {
    mockedApiGet.mockImplementation((url: string) =>
        Promise.resolve({ data: url.startsWith('/files/location/') ? location : summary })
    );
};

const renderViewer = () => {
    const activeImage = buildImageLibraryItem({ file_id: 7 });
    return render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <ImageViewerModal
                activeImage={activeImage}
                activeIndex={0}
                activeImageDate={null}
                dateFormatter={new Intl.DateTimeFormat('pt-BR')}
                filteredImages={[activeImage]}
                zoom={1}
                showDetails
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
        </QueryClientProvider>
    );
};

describe('ImageViewerModal details panel', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('shows GPS coordinates with an OpenStreetMap link opened in a new tab', async () => {
        mockEndpoints({ gps_latitude: -23.55, gps_longitude: -46.63 }, {});

        renderViewer();

        const mapLink = await screen.findByRole('link', { name: 'IMAGES_DETAIL_GPS_OPEN_MAP' });
        expect(screen.getByText('-23.550000, -46.630000')).toBeInTheDocument();
        expect(mapLink.getAttribute('href')).toContain(
            'https://www.openstreetmap.org/?mlat=-23.550000'
        );
        expect(mapLink).toHaveAttribute('target', '_blank');
        expect(mapLink.getAttribute('rel')).toContain('noopener');
    });

    it('shows no map link when the image has no GPS', async () => {
        mockEndpoints({ gps_latitude: null, gps_longitude: null }, {});

        renderViewer();

        await waitFor(() => expect(mockedApiGet).toHaveBeenCalledWith('/image/metadata/7'));
        expect(screen.getByText('IMAGES_DETAIL_GPS')).toBeInTheDocument();
        expect(screen.queryByRole('link', { name: 'IMAGES_DETAIL_GPS_OPEN_MAP' })).toBeNull();
    });

    it('shows caption, tags and OCR text from the metadata summary', async () => {
        mockEndpoints({ caption: 'A red car', tags: ['car', 'red'], ocr_text: 'PLATE 42' }, {});

        renderViewer();

        expect(await screen.findByText('A red car')).toBeInTheDocument();
        expect(screen.getAllByRole('button', { name: 'IMAGES_DETAIL_TAG_SEARCH' })).toHaveLength(2);
        expect(screen.getByRole('button', { name: 'IMAGES_DETAIL_OCR_SHOW' })).toBeInTheDocument();
        expect(screen.queryByText('PLATE 42')).not.toBeInTheDocument();
    });

    it('shows the location on disk and copies it', async () => {
        const writeText = jest.fn().mockResolvedValue(undefined);
        Object.defineProperty(navigator, 'clipboard', { value: { writeText }, configurable: true });
        mockEndpoints({}, { disk_path: 'D:\\photos\\Trip.jpg' });

        renderViewer();

        expect(await screen.findByText('D:\\photos\\Trip.jpg')).toBeInTheDocument();
        await act(async () => {
            fireEvent.click(
                screen.getByRole('button', {
                    name: 'IMAGES_DETAIL_COPY: IMAGES_DETAIL_DISK_LOCATION',
                })
            );
        });

        expect(writeText).toHaveBeenCalledWith('D:\\photos\\Trip.jpg');
    });
});
