import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';
import FileProvider from './index';
import { useFile } from './fileContext';
import { apiBase } from '@/service';
import { buildServerPage } from '@/shared/test/buildServerPage';

// Seam test: only apiBase is mocked; the real FileProvider + service/files.ts
// run, so each file-manager command asserts the exact endpoint/payload the
// backend files operation handlers decode.
jest.mock('@/service', () => ({
	apiBase: { get: jest.fn(), post: jest.fn(), put: jest.fn(), delete: jest.fn() },
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

const mockedApi = apiBase as unknown as {
	get: jest.Mock;
	post: jest.Mock;
	put: jest.Mock;
	delete: jest.Mock;
};

const sharedClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });

const wrapper = ({ children }: { children: ReactNode }) => {
	const client = sharedClient;
	return (
		<QueryClientProvider client={client}>
			<FileProvider>{children}</FileProvider>
		</QueryClientProvider>
	);
};

describe('features/files/fileProvider (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		sharedClient.clear();
		mockedApi.get.mockImplementation((url: string) => {
			if (url.startsWith('/files/recent/')) return Promise.resolve({ data: [] });
			return Promise.resolve({ data: { items: [], pagination: { has_next: false, page: 1 } } });
		});
		mockedApi.post.mockResolvedValue({ data: undefined });
		mockedApi.delete.mockResolvedValue({ data: undefined });
		mockedApi.put.mockResolvedValue({ data: { id: 5, starred: true } });
	});

	it('createFolder POSTs to /files/folder with name and parent_id', async () => {
		const { result } = renderHook(() => useFile(), { wrapper });

		await act(async () => {
			await result.current.createFolder('Nova', 7);
		});

		expect(mockedApi.post).toHaveBeenCalledWith('/files/folder', { name: 'Nova', parent_id: 7 });
	});

	it('promoteFileToHot POSTs to /tiering/promote/:id with no body', async () => {
		const { result } = renderHook(() => useFile(), { wrapper });

		await act(async () => {
			await result.current.promoteFileToHot(12);
		});

		expect(mockedApi.post).toHaveBeenCalledWith('/tiering/promote/12');
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

	it('deleteFile with permanent DELETEs /files/path with permanent=true as a query param', async () => {
		const { result } = renderHook(() => useFile(), { wrapper });

		await act(async () => {
			await result.current.deleteFile(5, true);
		});

		expect(mockedApi.delete).toHaveBeenCalledWith('/files/path', {
			data: { id: 5 },
			params: { permanent: true },
		});
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

	it('handleStarredItem PUTs the desired state to /files/starred/:id', async () => {
		const { result } = renderHook(() => useFile(), { wrapper });

		act(() => result.current.handleStarredItem(5));

		await waitFor(() =>
			expect(mockedApi.put).toHaveBeenCalledWith('/files/starred/5', { starred: true })
		);
	});

	it('handleStarredItem invalidates the home favorites view once settled', async () => {
		const invalidateQueries = jest.spyOn(sharedClient, 'invalidateQueries');
		const { result } = renderHook(() => useFile(), { wrapper });

		act(() => result.current.handleStarredItem(5));

		await waitFor(() =>
			expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['home', 'favorites'] })
		);
		invalidateQueries.mockRestore();
	});

	it('setFilesStarred POSTs ids and the starred flag to /files/starred', async () => {
		mockedApi.post.mockResolvedValue({
			data: { starred: true, updated_ids: [4, 5], not_found_ids: [] },
		});
		const { result } = renderHook(() => useFile(), { wrapper });

		let updatedIds: number[] = [];
		await act(async () => {
			updatedIds = await result.current.setFilesStarred([4, 5], true);
		});

		expect(mockedApi.post).toHaveBeenCalledWith('/files/starred', { ids: [4, 5], starred: true });
		expect(updatedIds).toEqual([4, 5]);
	});

	it('reindexOpenedFolder POSTs scope all when no folder is open', async () => {
		mockedApi.post.mockResolvedValue({ data: { job_id: 4 } });
		const { result } = renderHook(() => useFile(), { wrapper });

		await act(async () => {
			result.current.reindexOpenedFolder();
		});

		expect(mockedApi.post).toHaveBeenCalledWith('/indexing/reindex', { scope: 'all' });
	});

	it('exposes the items of every loaded page after fetchNextPage', async () => {
		const buildFile = (id: number) => ({ id, name: `file-${id}`, path: `/file-${id}`, type: 2 });
		mockedApi.get.mockImplementation((url: string, config?: { params?: { page?: number } }) => {
			if (url.startsWith('/files/recent/')) return Promise.resolve({ data: [] });
			if (config?.params?.page === 2) {
				return Promise.resolve({
					data: buildServerPage([buildFile(3)], { page: 2, pageSize: 60, hasNext: false }),
				});
			}
			return Promise.resolve({
				data: buildServerPage([buildFile(1), buildFile(2)], { page: 1, pageSize: 60, hasNext: true }),
			});
		});
		const { result } = renderHook(() => useFile(), { wrapper });

		await waitFor(() => expect(result.current.files).toHaveLength(2));
		expect(result.current.hasNextPage).toBe(true);

		act(() => result.current.fetchNextPage());

		await waitFor(() => expect(result.current.files).toHaveLength(3));
		expect(result.current.files.map((file) => file.id)).toEqual([1, 2, 3]);
		expect(mockedApi.get).toHaveBeenCalledWith('/files/tree', expect.objectContaining({ params: expect.objectContaining({ page: 2 }) }));
		expect(result.current.hasNextPage).toBe(false);
		expect(result.current.isFetchingNextPage).toBe(false);
	});
});
