import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { apiBase } from '@/service';
import { buildImageLibraryItem } from './imageLibraryTestFixtures';
import { useImageAlbumCards } from './useImageAlbumCards';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockedApiGet = apiBase.get as jest.Mock;

const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider
        client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
        {children}
    </QueryClientProvider>
);

describe('useImageAlbumCards', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('lists the five presets without calling the server while disabled', () => {
        const { result } = renderHook(() => useImageAlbumCards(false), { wrapper });

        expect(result.current.map((card) => card.id)).toEqual([
            'documents',
            'memes',
            'art',
            'landscapes',
            'portraits',
        ]);
        expect(result.current[0]).toEqual(
            expect.objectContaining({
                title: 'IMAGES_ALBUM_DOCUMENTS',
                description: 'IMAGES_ALBUM_DOCUMENTS_DESCRIPTION',
                imageCount: undefined,
                coverImageId: undefined,
            })
        );
        expect(mockedApiGet).not.toHaveBeenCalled();
    });

    it('fills each card with the server count and the first image as cover', async () => {
        mockedApiGet.mockImplementation(
            (url: string, config: { params: { category: string[] } }) => {
                const categories = config.params.category;
                if (url === '/image/library/count') {
                    return Promise.resolve({ data: { total: categories.length * 10 } });
                }
                return Promise.resolve({
                    data: {
                        items: [buildImageLibraryItem({ file_id: categories.length * 100 })],
                        next_cursor: '',
                        has_next: true,
                        page_size: 1,
                    },
                });
            }
        );

        const { result } = renderHook(() => useImageAlbumCards(true), { wrapper });

        await waitFor(() => expect(result.current[0]!.imageCount).toBe(20));
        await waitFor(() => expect(result.current[0]!.coverImageId).toBe(200));
        expect(result.current[1]!.imageCount).toBe(10);
        expect(mockedApiGet).toHaveBeenCalledWith(
            '/image/library',
            expect.objectContaining({
                params: expect.objectContaining({ category: ['meme'], page_size: 1 }),
            })
        );
    });
});
