import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import FilePathSync from './FilePathSync';

const mockUseFile = jest.fn();
const mockGetFileByPath = jest.fn();

jest.mock('@/components/providers/fileProvider/fileContext', () => ({
	__esModule: true,
	default: () => mockUseFile(),
}));

jest.mock('@/service/files', () => ({
	getFileByPath: (path: string) => mockGetFileByPath(path),
}));

const createQueryClient = () =>
	new QueryClient({
		defaultOptions: {
			queries: {
				retry: false,
			},
		},
	});

describe('FilePathSync', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockUseFile.mockReturnValue({
			selectResolvedItem: jest.fn(),
			selectedItem: null,
		});
	});

	it('selects the requested folder path once it is resolved', async () => {
		const selectResolvedItem = jest.fn();
		mockUseFile.mockReturnValue({
			selectResolvedItem,
			selectedItem: null,
		});
		mockGetFileByPath.mockResolvedValue({
			id: 42,
			name: 'travel',
			path: '/photos/travel',
			parent_path: '/photos',
			type: 1,
		});

		render(
			<QueryClientProvider client={createQueryClient()}>
				<MemoryRouter initialEntries={['/files?path=%2Fphotos%2Ftravel']}>
					<Routes>
						<Route path='/files' element={<FilePathSync />} />
					</Routes>
				</MemoryRouter>
			</QueryClientProvider>,
			);

			await waitFor(() => expect(mockGetFileByPath).toHaveBeenCalledWith('/photos/travel'));
			await waitFor(() => expect(selectResolvedItem).toHaveBeenCalledWith(expect.objectContaining({ id: 42 })));
		});

	it('ignores unknown paths without selecting a folder', async () => {
		const selectResolvedItem = jest.fn();
		mockUseFile.mockReturnValue({
			selectResolvedItem,
			selectedItem: null,
		});
		mockGetFileByPath.mockResolvedValue(null);

		render(
			<QueryClientProvider client={createQueryClient()}>
				<MemoryRouter initialEntries={['/files?path=%2Fphotos%2Fmissing']}>
					<Routes>
						<Route path='/files' element={<FilePathSync />} />
					</Routes>
				</MemoryRouter>
			</QueryClientProvider>,
		);

			await waitFor(() => expect(mockGetFileByPath).toHaveBeenCalledWith('/photos/missing'));
			expect(selectResolvedItem).not.toHaveBeenCalled();
		});

	it('does not reselect the folder when it is already active', async () => {
		const selectResolvedItem = jest.fn();
		mockUseFile.mockReturnValue({
			selectResolvedItem,
			selectedItem: { id: 42 },
		});
		mockGetFileByPath.mockResolvedValue({
			id: 42,
			name: 'travel',
			path: '/photos/travel',
			parent_path: '/photos',
			type: 1,
		});

		render(
			<QueryClientProvider client={createQueryClient()}>
				<MemoryRouter initialEntries={['/files?path=%2Fphotos%2Ftravel']}>
					<Routes>
						<Route path='/files' element={<FilePathSync />} />
					</Routes>
				</MemoryRouter>
			</QueryClientProvider>,
		);

			await waitFor(() => expect(mockGetFileByPath).toHaveBeenCalledWith('/photos/travel'));
			expect(selectResolvedItem).not.toHaveBeenCalled();
		});
});
