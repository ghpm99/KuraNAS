import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';
import FileProvider from './index';
import { useFile } from './fileContext';
import { apiBase } from '@/service';

// Seam test: only apiBase is mocked; the real FileProvider + service/files.ts
// run, so each file-manager command asserts the exact endpoint/payload the
// backend files operation handlers decode.
jest.mock('@/service', () => ({
	apiBase: { get: jest.fn(), post: jest.fn(), delete: jest.fn() },
}));

jest.mock('react-router-dom', () => ({
	...jest.requireActual('react-router-dom'),
	useLocation: () => ({ pathname: '/files', search: '', hash: '', state: null, key: 'default' }),
	useNavigate: () => jest.fn(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock; delete: jest.Mock };

const wrapper = ({ children }: { children: ReactNode }) => {
	const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	return (
		<QueryClientProvider client={client}>
			<FileProvider>{children}</FileProvider>
		</QueryClientProvider>
	);
};

describe('features/files/fileProvider (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.get.mockImplementation((url: string) => {
			if (url.startsWith('/files/recent/')) return Promise.resolve({ data: [] });
			return Promise.resolve({ data: { items: [], pagination: { has_next: false, page: 1 } } });
		});
		mockedApi.post.mockResolvedValue({ data: undefined });
		mockedApi.delete.mockResolvedValue({ data: undefined });
	});

	it('createFolder POSTs to /files/folder with name and parent_id', async () => {
		const { result } = renderHook(() => useFile(), { wrapper });

		await act(async () => {
			await result.current.createFolder('Nova', 7);
		});

		expect(mockedApi.post).toHaveBeenCalledWith('/files/folder', { name: 'Nova', parent_id: 7 });
	});

	it('renameFile POSTs to /files/rename with id and new_name', async () => {
		const { result } = renderHook(() => useFile(), { wrapper });

		await act(async () => {
			await result.current.renameFile(5, 'relatorio.pdf');
		});

		expect(mockedApi.post).toHaveBeenCalledWith('/files/rename', { id: 5, new_name: 'relatorio.pdf' });
	});

	it('moveFile POSTs to /files/move with the source and destination', async () => {
		const { result } = renderHook(() => useFile(), { wrapper });

		await act(async () => {
			await result.current.moveFile(3, 9, '/data/dst');
		});

		expect(mockedApi.post).toHaveBeenCalledWith(
			'/files/move',
			expect.objectContaining({ source_id: 3, destination_folder_id: 9 })
		);
	});

	it('deleteFile DELETEs /files/path with the id in the body', async () => {
		const { result } = renderHook(() => useFile(), { wrapper });

		await act(async () => {
			await result.current.deleteFile(5);
		});

		expect(mockedApi.delete).toHaveBeenCalledWith('/files/path', { data: { id: 5 } });
	});

	it('copyFile POSTs to /files/copy with the source, destination and new name', async () => {
		const { result } = renderHook(() => useFile(), { wrapper });

		await act(async () => {
			await result.current.copyFile(3, 9, '/data/dst', 'copia.pdf');
		});

		expect(mockedApi.post).toHaveBeenCalledWith(
			'/files/copy',
			expect.objectContaining({ source_id: 3, destination_folder_id: 9, new_name: 'copia.pdf' })
		);
	});

	it('handleStarredItem POSTs to /files/starred/:id', async () => {
		const { result } = renderHook(() => useFile(), { wrapper });

		act(() => result.current.handleStarredItem(5));

		await waitFor(() => expect(mockedApi.post).toHaveBeenCalledWith('/files/starred/5'));
	});

	it('rescanFiles POSTs the manual rescan to /files/update', async () => {
		const { result } = renderHook(() => useFile(), { wrapper });

		await act(async () => {
			await result.current.rescanFiles();
		});

		expect(mockedApi.post).toHaveBeenCalledWith(
			'/files/update',
			expect.any(FormData),
			expect.objectContaining({ headers: { 'Content-Type': 'multipart/form-data' } })
		);
	});

	it('exposes the items of every loaded page after fetchNextPage', async () => {
		const buildFile = (id: number) => ({ id, name: `file-${id}`, path: `/file-${id}`, type: 2 });
		mockedApi.get.mockImplementation((url: string, config?: { params?: { page?: number } }) => {
			if (url.startsWith('/files/recent/')) return Promise.resolve({ data: [] });
			if (config?.params?.page === 2) {
				return Promise.resolve({
					data: { items: [buildFile(3)], pagination: { hasNext: false, page: 2 } },
				});
			}
			return Promise.resolve({
				data: { items: [buildFile(1), buildFile(2)], pagination: { hasNext: true, page: 1 } },
			});
		});
		const { result } = renderHook(() => useFile(), { wrapper });

		await waitFor(() => expect(result.current.files).toHaveLength(2));
		expect(result.current.hasNextPage).toBe(true);

		act(() => result.current.fetchNextPage());

		await waitFor(() => expect(result.current.files).toHaveLength(3));
		expect(result.current.files.map((file) => file.id)).toEqual([1, 2, 3]);
		expect(result.current.hasNextPage).toBe(false);
		expect(result.current.isFetchingNextPage).toBe(false);
	});
});
