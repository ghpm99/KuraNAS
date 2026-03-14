import { fireEvent, render, screen } from '@testing-library/react';
import FilesExplorerScreen from './FilesExplorerScreen';

const mockUseFile = jest.fn();

jest.mock('@/components/providers/fileProvider/fileContext', () => ({
	__esModule: true,
	default: () => mockUseFile(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

jest.mock('@/components/actionBar', () => () => <div>ActionBarMock</div>);
jest.mock('@/components/fileContent', () => ({ viewMode, showHeading }: any) => (
	<div data-testid='file-content' data-view-mode={viewMode} data-show-heading={String(showHeading)}>
		FileContentMock
	</div>
));
jest.mock('@/components/fileDetails', () => () => <div>FileDetailsMock</div>);
jest.mock('@/components/layout/Sidebar/components/folderTree', () => () => <div>FolderTreeMock</div>);
jest.mock('@/components/tabs', () => () => <div>TabsMock</div>);

describe('FilesExplorerScreen', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockUseFile.mockReturnValue({
			files: [
				{
					id: 1,
					name: 'media',
					path: '/media',
					parent_path: '/',
					type: 1,
					file_children: [
						{
							id: 2,
							name: 'movie.mp4',
							path: '/media/movie.mp4',
							parent_path: '/media',
							type: 2,
							format: '.mp4',
							size: 42,
						},
					],
				},
			],
			selectedItem: null,
			handleSelectItem: jest.fn(),
			fileListFilter: 'all',
		});
	});

	it('renders explorer structure and switches view mode', () => {
		render(<FilesExplorerScreen />);

		expect(screen.getByText('FILES_PAGE_TITLE')).toBeInTheDocument();
		expect(screen.getByText('FILES_PAGE_DESCRIPTION')).toBeInTheDocument();
		expect(screen.getByText('FILES_EXPLORER_EYEBROW')).toBeInTheDocument();
		expect(screen.getByTestId('file-content')).toHaveAttribute('data-view-mode', 'grid');
		expect(screen.getByTestId('file-content')).toHaveAttribute('data-show-heading', 'false');

		fireEvent.click(screen.getByRole('button', { name: 'FILES_VIEW_LIST' }));
		expect(screen.getByTestId('file-content')).toHaveAttribute('data-view-mode', 'list');
	});

	it('renders breadcrumb and preview when a file is selected', () => {
		mockUseFile.mockReturnValue({
			...mockUseFile(),
			selectedItem: {
				id: 2,
				name: 'movie.mp4',
				path: '/media/movie.mp4',
				parent_path: '/media',
				type: 2,
				format: '.mp4',
				size: 42,
			},
		});

		render(<FilesExplorerScreen />);

		expect(screen.getByText('media')).toBeInTheDocument();
		expect(screen.getAllByText('movie.mp4').length).toBeGreaterThan(0);
		expect(screen.getByText('FileDetailsMock')).toBeInTheDocument();
	});

	it('opens the tree drawer action', () => {
		render(<FilesExplorerScreen />);

		fireEvent.click(screen.getByRole('button', { name: 'FILES_OPEN_TREE' }));
		expect(screen.getByText('FolderTreeMock')).toBeInTheDocument();
	});
});
