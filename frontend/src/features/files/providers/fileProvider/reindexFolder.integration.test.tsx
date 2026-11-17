import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';
import FileProvider from './index';
import { useFile } from './fileContext';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
	apiBase: { get: jest.fn(), post: jest.fn(), put: jest.fn(), delete: jest.fn() },
}));

jest.mock('react-router-dom', () => ({
	...jest.requireActual('react-router-dom'),
	useLocation: () => ({
		pathname: '/files/media/Photos',
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

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock };

const openedFolder = { id: 12, name: 'Photos', path: '/media/Photos', type: 1 };
const emptyPage = { items: [], pagination: { has_next: false, page: 1 } };

const buildWrapper = () => {
	const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	return ({ children }: { children: ReactNode }) => (
		<QueryClientProvider client={client}>
			<FileProvider>{children}</FileProvider>
		</QueryClientProvider>
	);
};

const countCalls = (url: string) =>
	mockedApi.get.mock.calls.filter(([calledUrl]) => calledUrl === url).length;

describe('features/files/fileProvider reindex folder (seam)', () => {
	let jobStatus: string;

	beforeEach(() => {
		jest.clearAllMocks();
		jobStatus = 'running';
		mockedApi.get.mockImplementation((url: string) => {
			if (url === '/files/path') return Promise.resolve({ data: { items: [openedFolder] } });
			if (url === '/jobs/5') {
				return Promise.resolve({
					data: {
						id: 5,
						type: 'reindex',
						status: jobStatus,
						progress: { progress: 40 },
					},
				});
			}
			if (url.startsWith('/files/recent/') || url.startsWith('/files/ancestors/')) return Promise.resolve({ data: [] });
			return Promise.resolve({ data: emptyPage });
		});
		mockedApi.post.mockResolvedValue({ data: { job_id: 5 } });
	});

	it('POSTs the opened folder id and reloads the listing only after the job completes', async () => {
		const { result } = renderHook(() => useFile(), { wrapper: buildWrapper() });
		await waitFor(() => expect(result.current.selectedItem?.id).toBe(12));
		await waitFor(() => expect(countCalls('/files/tree')).toBeGreaterThan(0));

		act(() => {
			result.current.reindexOpenedFolder();
		});

		await waitFor(() => expect(countCalls('/jobs/5')).toBeGreaterThan(0));
		expect(mockedApi.post).toHaveBeenCalledWith('/indexing/reindex', {
			scope: 'folder',
			folder_id: 12,
		});
		expect(result.current.isReindexing).toBe(true);
		const listingCallsWhileRunning = countCalls('/files/tree');

		jobStatus = 'completed';

		await waitFor(() => expect(result.current.isReindexing).toBe(false), { timeout: 5000 });
		await waitFor(() =>
			expect(countCalls('/files/tree')).toBeGreaterThan(listingCallsWhileRunning)
		);
	});
});
