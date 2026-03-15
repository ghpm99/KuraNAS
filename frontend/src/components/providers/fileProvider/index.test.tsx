import { act, render, screen, waitFor } from '@testing-library/react';
import FileProvider from './index';
import { useFile } from './fileContext';
import { useInfiniteQuery, useMutation, useQuery } from '@tanstack/react-query';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
	apiBase: {
		get: jest.fn(),
		post: jest.fn(),
	},
}));

jest.mock('@tanstack/react-query', () => ({
	useInfiniteQuery: jest.fn(),
	useQuery: jest.fn(),
	useMutation: jest.fn(),
}));

const mockedUseInfiniteQuery = useInfiniteQuery as jest.Mock;
const mockedUseQuery = useQuery as jest.Mock;
const mockedUseMutation = useMutation as jest.Mock;
const mockedApiGet = apiBase.get as jest.Mock;
const mockedApiPost = apiBase.post as jest.Mock;
const mockRefetch = jest.fn();

	function Consumer() {
		const ctx = useFile();
	return (
		<div>
				<span data-testid="files-count">{ctx.files.length}</span>
				<span data-testid="selected-id">{ctx.selectedItem?.id ?? 'none'}</span>
				<span data-testid="expanded">{ctx.expandedItems.join(',')}</span>
				<button onClick={() => ctx.selectResolvedItem({ id: 3, type: 1, name: 'external', file_children: [] } as any)}>resolve</button>
				<button onClick={() => ctx.handleSelectItem(1)}>select-folder</button>
			<button onClick={() => ctx.handleSelectItem(2)}>select-file</button>
			<button onClick={() => ctx.handleSelectItem(null)}>clear</button>
			<button onClick={() => ctx.handleStarredItem(2)}>star</button>
		</div>
	);
}

describe('providers/fileProvider/index', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApiGet.mockResolvedValue({ data: [] });
		mockedApiPost.mockResolvedValue({});

		mockedUseInfiniteQuery.mockReturnValue({
			status: 'success',
			data: {
				pages: [
					{
						items: [
							{ id: 1, type: 1, name: 'folder', file_children: [{ id: 2, type: 2, name: 'file' }] },
						],
						pagination: { hasNext: false, page: 1 },
					},
				],
			},
			refetch: mockRefetch,
		});

		mockedUseQuery.mockReturnValue({ data: [], isLoading: false });
		mockedUseMutation.mockImplementation(({ mutationFn, onSuccess }: any) => ({
			mutate: async (value: unknown) => {
				await mutationFn(value);
				onSuccess?.();
			},
		}));
	});

	it('provides tree data and selection/expansion behavior', async () => {
		render(
			<FileProvider>
				<Consumer />
			</FileProvider>,
		);

		await waitFor(() => {
			expect(screen.getByTestId('files-count')).toHaveTextContent('1');
			expect(screen.getByTestId('selected-id')).toHaveTextContent('none');
		});

			act(() => {
				screen.getByText('resolve').click();
			});
			expect(screen.getByTestId('selected-id')).toHaveTextContent('3');

			act(() => {
				screen.getByText('select-folder').click();
			});
		expect(screen.getByTestId('selected-id')).toHaveTextContent('1');
		expect(screen.getByTestId('expanded')).toHaveTextContent('1');

		// Clicking the same item again should toggle deselect
		act(() => {
			screen.getByText('select-folder').click();
		});
		expect(screen.getByTestId('selected-id')).toHaveTextContent('none');

		act(() => {
			screen.getByText('clear').click();
		});
		expect(screen.getByTestId('selected-id')).toHaveTextContent('none');
	});

	it('runs mutation when starring item', async () => {
		render(
			<FileProvider>
				<Consumer />
			</FileProvider>,
		);

		await waitFor(() => {
			expect(screen.getByTestId('files-count')).toHaveTextContent('1');
		});

		await act(async () => {
			screen.getByText('star').click();
		});

		expect(mockedApiPost).toHaveBeenCalledWith('/files/starred/2');
		expect(mockRefetch).toHaveBeenCalled();
	});

	it('executes query functions for tree and recent access branches', async () => {
		render(
			<FileProvider>
				<Consumer />
			</FileProvider>,
		);

		await waitFor(() => {
			expect(screen.getByTestId('files-count')).toHaveTextContent('1');
		});

		let infiniteOptions = mockedUseInfiniteQuery.mock.calls[0][0];
		await infiniteOptions.queryFn({ pageParam: 3 });
		expect(mockedApiGet).toHaveBeenCalledWith('/files/tree', {
			params: { page_size: 200, file_parent: undefined, page: 3, category: 'all' },
		});
		await infiniteOptions.queryFn({});
		expect(mockedApiGet).toHaveBeenCalledWith('/files/tree', {
			params: { page_size: 200, file_parent: undefined, page: 1, category: 'all' },
		});
		expect(infiniteOptions.getNextPageParam({ pagination: { hasNext: true, page: 5 } })).toBe(6);
		expect(infiniteOptions.getNextPageParam({ pagination: { hasNext: false, page: 5 } })).toBeUndefined();

		let recentOptions = mockedUseQuery.mock.calls[0][0];
		await expect(recentOptions.queryFn()).resolves.toEqual([]);

		act(() => {
			screen.getByText('select-file').click();
		});
		recentOptions = mockedUseQuery.mock.calls[mockedUseQuery.mock.calls.length - 1][0];
		mockedApiGet.mockResolvedValueOnce({ data: [{ id: 10 }] });
		await expect(recentOptions.queryFn()).resolves.toEqual([{ id: 10 }]);
		expect(mockedApiGet).toHaveBeenCalledWith('/files/recent/2');
	});
});
