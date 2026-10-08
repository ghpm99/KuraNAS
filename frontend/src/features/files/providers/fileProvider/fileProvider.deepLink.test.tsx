import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';
import FileProvider from './index';
import { useFile } from './fileContext';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn(), delete: jest.fn() },
}));

jest.mock('react-router-dom', () => ({
    ...jest.requireActual('react-router-dom'),
    useLocation: () => ({
        pathname: '/files/fotos/2024',
        search: '',
        hash: '',
        state: null,
        key: 'default',
    }),
    useNavigate: () => jest.fn(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock };

const folder = (id: number, name: string, path: string, parentPath: string) => ({
    id,
    name,
    path,
    parent_path: parentPath,
    type: 1,
    directory_content_count: 1,
});

const page = (items: unknown[]) => ({ items, pagination: { hasNext: false, page: 1 } });

const wrapper = ({ children }: { children: ReactNode }) => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return (
        <QueryClientProvider client={client}>
            <FileProvider>{children}</FileProvider>
        </QueryClientProvider>
    );
};

const mockDeepLinkBackend = (ancestorsResponse: () => Promise<unknown>) => {
    mockedApi.get.mockImplementation((url: string, config?: { params?: { file_parent?: number } }) => {
        if (url === '/files/path') {
            return Promise.resolve({ data: page([folder(3, '2024', '/fotos/2024', '/fotos')]) });
        }
        if (url === '/files/ancestors/3') return ancestorsResponse();
        if (url === '/files/tree') {
            const parentId = config?.params?.file_parent;
            if (parentId === 3) return Promise.resolve({ data: page([]) });
            if (parentId === 2) {
                return Promise.resolve({ data: page([folder(3, '2024', '/fotos/2024', '/fotos')]) });
            }
            return Promise.resolve({ data: page([folder(2, 'fotos', '/fotos', '/')]) });
        }
        return Promise.resolve({ data: [] });
    });
};

describe('FileProvider deep link tree expansion (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('requests the ancestors by id and expands the tree down to the opened folder', async () => {
        mockDeepLinkBackend(() =>
            Promise.resolve({
                data: [
                    { id: 1, name: 'main', path: '/', type: 1 },
                    { id: 2, name: 'fotos', path: '/fotos', type: 1 },
                ],
            })
        );

        const { result } = renderHook(() => useFile(), { wrapper });

        await waitFor(() => {
            expect(result.current.files[0]?.file_children?.[0]?.id).toBe(3);
        });

        expect(mockedApi.get).toHaveBeenCalledWith('/files/ancestors/3');
        expect(result.current.files[0]?.id).toBe(2);
        expect(result.current.selectedItem?.id).toBe(3);
        expect(result.current.expandedItems).toEqual([2, 3]);
    });

    it('keeps the level zero tree usable when the ancestors request fails', async () => {
        mockDeepLinkBackend(() => Promise.reject(new Error('ancestors down')));

        const { result } = renderHook(() => useFile(), { wrapper });

        await waitFor(() => {
            expect(result.current.files.map((node) => node.id)).toEqual([2]);
        });

        expect(result.current.selectedItem?.id).toBe(3);
    });
});
