import { act, renderHook } from '@testing-library/react';
import useFavoritesScreen from './useFavoritesScreen';

const mockUseFile = jest.fn();

jest.mock('@/components/providers/fileProvider/fileContext', () => ({
	__esModule: true,
	default: () => mockUseFile(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

describe('useFavoritesScreen', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('filters starred scope by folders, files, and media', () => {
		mockUseFile.mockReturnValue({
			files: [
				{
					id: 1,
					name: 'Projects',
					path: '/projects',
					parent_path: '/',
					type: 1,
					format: '',
					size: 0,
					updated_at: '',
					created_at: '',
					deleted_at: '',
					last_interaction: '',
					last_backup: '',
					check_sum: '',
					directory_content_count: 2,
					starred: true,
				},
				{
					id: 2,
					name: 'Invoice.pdf',
					path: '/docs/invoice.pdf',
					parent_path: '/docs',
					type: 2,
					format: '.pdf',
					size: 42,
					updated_at: '',
					created_at: '',
					deleted_at: '',
					last_interaction: '',
					last_backup: '',
					check_sum: '',
					directory_content_count: 0,
					starred: true,
				},
				{
					id: 3,
					name: 'Song.mp3',
					path: '/music/song.mp3',
					parent_path: '/music',
					type: 2,
					format: '.mp3',
					size: 84,
					updated_at: '',
					created_at: '',
					deleted_at: '',
					last_interaction: '',
					last_backup: '',
					check_sum: '',
					directory_content_count: 0,
					starred: true,
				},
			],
			selectedItem: null,
			handleSelectItem: jest.fn(),
		});

		const { result } = renderHook(() => useFavoritesScreen());

		expect(result.current.currentTitle).toBe('STARRED_FILES');
		expect(result.current.contextPath).toBe('FAVORITES_CONTEXT_ALL');
		expect(result.current.itemCountLabel).toBe('3 ITENS');
		expect(result.current.filterOptions).toEqual([
			{ value: 'all', label: 'FAVORITES_FILTER_ALL', count: 3 },
			{ value: 'folders', label: 'FAVORITES_FILTER_FOLDERS', count: 1 },
			{ value: 'files', label: 'FAVORITES_FILTER_FILES', count: 1 },
			{ value: 'media', label: 'FAVORITES_FILTER_MEDIA', count: 1 },
		]);

		act(() => {
			result.current.setActiveFilter('folders');
		});
		expect(result.current.filteredItems.map((item) => item.name)).toEqual(['Projects']);
		expect(result.current.activeFilterLabel).toBe('FAVORITES_FILTER_FOLDERS');

		act(() => {
			result.current.setActiveFilter('files');
		});
		expect(result.current.filteredItems.map((item) => item.name)).toEqual(['Invoice.pdf']);

		act(() => {
			result.current.setActiveFilter('media');
		});
		expect(result.current.filteredItems.map((item) => item.name)).toEqual(['Song.mp3']);
	});

	it('uses the selected folder as the active scope and keeps breadcrumb trail', () => {
		const folder = {
			id: 10,
			name: 'Captured',
			path: '/captured',
			parent_path: '/',
			type: 1,
			format: '',
			size: 0,
			updated_at: '',
			created_at: '',
			deleted_at: '',
			last_interaction: '',
			last_backup: '',
			check_sum: '',
			directory_content_count: 2,
			starred: true,
			file_children: [
				{
					id: 11,
					name: 'frame.jpg',
					path: '/captured/frame.jpg',
					parent_path: '/captured',
					type: 2,
					format: '.jpg',
					size: 30,
					updated_at: '',
					created_at: '',
					deleted_at: '',
					last_interaction: '',
					last_backup: '',
					check_sum: '',
					directory_content_count: 0,
					starred: true,
				},
				{
					id: 12,
					name: 'notes.txt',
					path: '/captured/notes.txt',
					parent_path: '/captured',
					type: 2,
					format: '.txt',
					size: 12,
					updated_at: '',
					created_at: '',
					deleted_at: '',
					last_interaction: '',
					last_backup: '',
					check_sum: '',
					directory_content_count: 0,
					starred: true,
				},
			],
		};

		mockUseFile.mockReturnValue({
			files: [folder],
			selectedItem: folder,
			handleSelectItem: jest.fn(),
		});

		const { result } = renderHook(() => useFavoritesScreen());

		expect(result.current.currentTitle).toBe('Captured');
		expect(result.current.contextPath).toBe('/captured');
		expect(result.current.breadcrumbSegments.map((segment) => segment.label)).toEqual(['STARRED_FILES', 'Captured']);

		act(() => {
			result.current.setActiveFilter('media');
		});

		expect(result.current.filteredItems.map((item) => item.name)).toEqual(['frame.jpg']);
		expect(result.current.itemCountLabel).toBe('1 ITEM');
	});
});
