import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { imageCategories } from '@/types/imageLibrary';
import { apiBase } from '@/service';
import { buildImageLibraryItem } from '../imageLibraryTestFixtures';
import { useImageViewerModal } from './useImageViewerModal';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string>) =>
            key === 'IMAGES_VIEWER_POSITION' ? `${params?.current} de ${params?.total}` : key,
    }),
}));

const mockedApiGet = apiBase.get as jest.Mock;
const dateFormatter = new Intl.DateTimeFormat('pt-BR', { dateStyle: 'medium', timeStyle: 'short' });

const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider
        client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
        {children}
    </QueryClientProvider>
);

const findValue = (sections: ReturnType<typeof useImageViewerModal>['details'], label: string) =>
    sections.flatMap((section) => section.items).find((item) => item.label === label)?.value;

describe('useImageViewerModal', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('falls back to not-available values while the metadata summary is missing', () => {
        mockedApiGet.mockRejectedValue(new Error('offline'));

        const { result } = renderHook(
            () =>
                useImageViewerModal({
                    activeImage: buildImageLibraryItem({ width: 0, height: 0, format: '' }),
                    activeImageDate: null,
                    activeIndex: 0,
                    totalImages: 1,
                    dateFormatter,
                }),
            { wrapper }
        );

        expect(findValue(result.current.details, 'IMAGES_DETAIL_DIMENSIONS')).toBe(
            'COMMON_NOT_AVAILABLE'
        );
        expect(findValue(result.current.details, 'IMAGES_DETAIL_FORMAT')).toBe(
            'COMMON_NOT_AVAILABLE'
        );
        expect(findValue(result.current.details, 'IMAGES_DETAIL_DATE')).toBe(
            'IMAGES_DATE_UNAVAILABLE'
        );
        expect(findValue(result.current.details, 'IMAGES_DETAIL_CAMERA')).toBe(
            'COMMON_NOT_AVAILABLE'
        );
    });

    it('merges library fields with the EXIF summary fetched from the metadata endpoint', async () => {
        mockedApiGet.mockResolvedValue({
            data: {
                width: 1600,
                height: 900,
                make: 'Sony',
                model: 'A7',
                lens_model: '24-70mm',
                datetime_original: '',
                exposure_time: 0.008,
                f_number: 2.8,
                iso: 400,
                focal_length: 35,
            },
        });

        const { result } = renderHook(
            () =>
                useImageViewerModal({
                    activeImage: buildImageLibraryItem({ width: 0, height: 0 }),
                    activeImageDate: new Date('2026-03-10T10:00:00Z'),
                    activeIndex: 2,
                    totalImages: 10,
                    dateFormatter,
                }),
            { wrapper }
        );

        await waitFor(() =>
            expect(findValue(result.current.details, 'IMAGES_DETAIL_CAMERA')).toBe('Sony A7')
        );
        expect(mockedApiGet).toHaveBeenCalledWith('/image/metadata/7');
        expect(result.current.folderPath).toBe('/photos/travel');
        expect(result.current.positionLabel).toBe('3 de 10');
        expect(findValue(result.current.details, 'IMAGES_DETAIL_DIMENSIONS')).toBe('1600 x 900');
        expect(findValue(result.current.details, 'IMAGES_DETAIL_LENS')).toBe('24-70mm');
        expect(findValue(result.current.details, 'IMAGES_DETAIL_ISO')).toBe('400');
        expect(findValue(result.current.details, 'IMAGES_DETAIL_FOCAL')).toBe('35mm');
        expect(findValue(result.current.details, 'IMAGES_DETAIL_APERTURE')).toBe('f/2.8');
        expect(findValue(result.current.details, 'IMAGES_DETAIL_EXPOSURE')).toBe('1/125s');
    });

    it('formats long exposures in seconds', async () => {
        mockedApiGet.mockResolvedValue({ data: { exposure_time: 2 } });

        const { result } = renderHook(
            () =>
                useImageViewerModal({
                    activeImage: buildImageLibraryItem(),
                    activeImageDate: null,
                    activeIndex: 0,
                    totalImages: 1,
                    dateFormatter,
                }),
            { wrapper }
        );

        await waitFor(() =>
            expect(findValue(result.current.details, 'IMAGES_DETAIL_EXPOSURE')).toBe('2s')
        );
    });

    it.each(imageCategories)('resolves a translated label for category %s', (category) => {
        mockedApiGet.mockRejectedValue(new Error('offline'));

        const { result } = renderHook(
            () =>
                useImageViewerModal({
                    activeImage: buildImageLibraryItem({ category }),
                    activeImageDate: null,
                    activeIndex: 0,
                    totalImages: 1,
                    dateFormatter,
                }),
            { wrapper }
        );

        const categoryLabel = findValue(result.current.details, 'IMAGES_DETAIL_CATEGORY');
        expect(categoryLabel).toMatch(/^IMAGES_CLASSIFICATION_/);
        expect(categoryLabel).not.toBe('undefined');
    });

    it('never asks for an undefined label when the server sends an unknown category', () => {
        mockedApiGet.mockRejectedValue(new Error('offline'));

        const { result } = renderHook(
            () =>
                useImageViewerModal({
                    activeImage: buildImageLibraryItem({
                        category: 'brand_new' as never,
                    }),
                    activeImageDate: null,
                    activeIndex: 0,
                    totalImages: 1,
                    dateFormatter,
                }),
            { wrapper }
        );

        expect(findValue(result.current.details, 'IMAGES_DETAIL_CATEGORY')).toBe(
            'IMAGES_CLASSIFICATION_OTHER'
        );
    });

    describe('extended details', () => {
        const mockEndpoints = (
            summary: Record<string, unknown>,
            location?: Record<string, unknown>
        ) => {
            mockedApiGet.mockImplementation((url: string) => {
                if (url.startsWith('/files/location/')) {
                    return location
                        ? Promise.resolve({ data: location })
                        : Promise.reject(new Error('offline'));
                }
                return Promise.resolve({ data: summary });
            });
        };

        const renderDetails = (totalImages: number | null = 1) =>
            renderHook(
                () =>
                    useImageViewerModal({
                        activeImage: buildImageLibraryItem(),
                        activeImageDate: null,
                        activeIndex: 0,
                        totalImages,
                        dateFormatter,
                    }),
                { wrapper }
            );

        const findItem = (
            sections: ReturnType<typeof useImageViewerModal>['details'],
            label: string
        ) => sections.flatMap((section) => section.items).find((item) => item.label === label);

        it('shows the coordinates with a user-clicked OpenStreetMap link', async () => {
            mockEndpoints({ gps_latitude: -23.55, gps_longitude: -46.63 });

            const { result } = renderDetails();

            await waitFor(() =>
                expect(findItem(result.current.details, 'IMAGES_DETAIL_GPS')?.link).toBeDefined()
            );
            const gpsItem = findItem(result.current.details, 'IMAGES_DETAIL_GPS');
            expect(gpsItem?.value).toBe('-23.550000, -46.630000');
            expect(gpsItem?.link?.href).toBe(
                'https://www.openstreetmap.org/?mlat=-23.550000&mlon=-46.630000#map=15/-23.550000/-46.630000'
            );
            expect(gpsItem?.link?.label).toBe('IMAGES_DETAIL_GPS_OPEN_MAP');
        });

        it.each([
            ['absent', {}],
            ['null', { gps_latitude: null, gps_longitude: null }],
            ['exact zero zero', { gps_latitude: 0, gps_longitude: 0 }],
            ['only latitude', { gps_latitude: 10 }],
            ['not finite', { gps_latitude: Number.NaN, gps_longitude: 10 }],
        ])('shows no link when the GPS is %s', async (_label, summary) => {
            mockEndpoints(summary);

            const { result } = renderDetails();

            await waitFor(() => expect(mockedApiGet).toHaveBeenCalledWith('/image/metadata/7'));
            const gpsItem = findItem(result.current.details, 'IMAGES_DETAIL_GPS');
            expect(gpsItem?.value).toBe('COMMON_NOT_AVAILABLE');
            expect(gpsItem?.link).toBeUndefined();
        });

        it('keeps a real coordinate that has a zero axis', async () => {
            mockEndpoints({ gps_latitude: 0, gps_longitude: 12.5 });

            const { result } = renderDetails();

            await waitFor(() =>
                expect(findItem(result.current.details, 'IMAGES_DETAIL_GPS')?.link).toBeDefined()
            );
        });

        it('exposes the disk location with a copy value from the location endpoint', async () => {
            mockEndpoints({}, { disk_path: 'D:\\photos\\Trip.jpg' });

            const { result } = renderDetails();

            await waitFor(() =>
                expect(findItem(result.current.details, 'IMAGES_DETAIL_DISK_LOCATION')?.value).toBe(
                    'D:\\photos\\Trip.jpg'
                )
            );
            expect(mockedApiGet).toHaveBeenCalledWith('/files/location/7');
            expect(findItem(result.current.details, 'IMAGES_DETAIL_DISK_LOCATION')?.copyValue).toBe(
                'D:\\photos\\Trip.jpg'
            );
        });

        it('has no copy value while the location is unavailable', () => {
            mockEndpoints({});

            const { result } = renderDetails();

            const locationItem = findItem(result.current.details, 'IMAGES_DETAIL_DISK_LOCATION');
            expect(locationItem?.value).toBe('COMMON_NOT_AVAILABLE');
            expect(locationItem?.copyValue).toBeUndefined();
        });

        it('shows software, description, classification and the capture date from the summary', async () => {
            mockEndpoints({
                software: 'Lightroom',
                image_description: 'Sunset',
                taken_at: '2026-03-10T10:00:00Z',
                classification_confidence: 0.834,
                suggested_name: 'sunset-beach',
            });

            const { result } = renderDetails();

            await waitFor(() =>
                expect(findValue(result.current.details, 'IMAGES_DETAIL_SOFTWARE')).toBe(
                    'Lightroom'
                )
            );
            expect(findValue(result.current.details, 'IMAGES_DETAIL_DESCRIPTION')).toBe('Sunset');
            expect(findValue(result.current.details, 'IMAGES_DETAIL_CONFIDENCE')).toBe('83%');
            expect(findValue(result.current.details, 'IMAGES_DETAIL_SUGGESTED_NAME')).toBe(
                'sunset-beach'
            );
            expect(findValue(result.current.details, 'IMAGES_DETAIL_DATE')).toBe(
                dateFormatter.format(new Date('2026-03-10T10:00:00Z'))
            );
        });

        it('omits the position label when the total is unknown', () => {
            mockEndpoints({});

            const { result } = renderDetails(null);

            expect(result.current.positionLabel).toBe('');
        });
    });
});
