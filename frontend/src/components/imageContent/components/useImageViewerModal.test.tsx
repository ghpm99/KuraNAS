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
});
