import { render, screen } from '@testing-library/react';
import { fireEvent } from '@testing-library/react';
import FileContent from './fileContent';

const mockUseFile = jest.fn();
const mockOpenMediaItem = jest.fn();

jest.mock('../providers/fileProvider/fileContext', () => ({
	__esModule: true,
	default: () => mockUseFile(),
}));
jest.mock('@/components/hooks/useMediaOpener/useMediaOpener', () => ({
	__esModule: true,
	default: () => ({
		openMediaItem: (...args: any[]) => mockOpenMediaItem(...args),
	}),
}));

jest.mock('../i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (k: string) => k }),
}));

jest.mock('../fileCard', () => ({ title, metadata, onClick, onClickStar }: any) => (
	<div>
		<button onClick={onClick}>{title}</button>
		<button onClick={onClickStar}>star-{title}</button>
		<span>{metadata}</span>
	</div>
));
jest.mock('./components/fileViewer/fileViewer', () => ({ file }: any) => <div>viewer:{file.name}</div>);

describe('fileContent', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockOpenMediaItem.mockReturnValue(false);
	});

	it('renders pending and error states', () => {
		mockUseFile.mockReturnValue({ status: 'pending', selectedItem: null, files: [] });
		render(<FileContent />);
		expect(screen.getByText('LOADING')).toBeInTheDocument();

		mockUseFile.mockReturnValue({ status: 'error', selectedItem: null, files: [] });
		render(<FileContent />);
		expect(screen.getByText('ERROR_LOADING_FILES')).toBeInTheDocument();
	});

	it('renders root files, directory and file preview branches', () => {
		const rootHandleSelectItem = jest.fn();
		const rootHandleStarredItem = jest.fn();
		mockUseFile.mockReturnValue({
			status: 'success',
			handleSelectItem: rootHandleSelectItem,
			handleStarredItem: rootHandleStarredItem,
			selectedItem: null,
			files: [
				{ id: 1, type: 2, name: 'song', format: '.mp3', size: 1024, starred: false },
				{ id: 4, type: 1, name: 'docs', directory_content_count: 1, starred: false },
			],
		});
		render(<FileContent />);
		expect(screen.getByText('FILES')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'song' })).toBeInTheDocument();
		expect(screen.getByText(/FOLDER - 1 ITEM/)).toBeInTheDocument();
		fireEvent.click(screen.getByRole('button', { name: 'song' }));
		expect(mockOpenMediaItem).toHaveBeenCalledWith(expect.objectContaining({ id: 1, name: 'song' }));
		expect(rootHandleSelectItem).toHaveBeenCalledWith(1);
		fireEvent.click(screen.getByRole('button', { name: 'star-song' }));
		expect(rootHandleStarredItem).toHaveBeenCalledWith(1);

		const handleSelectItem = jest.fn();
		const handleStarredItem = jest.fn();
		mockUseFile.mockReturnValue({
			status: 'success',
			handleSelectItem,
			handleStarredItem,
			selectedItem: {
				id: 2,
				type: 1,
				name: 'folder',
				file_children: [
					{ id: 3, type: 2, name: 'child', format: '.txt', size: 50, starred: true },
					{ id: 5, type: 1, name: 'subfolder', directory_content_count: 3, starred: false },
				],
			},
			files: [],
		});
		render(<FileContent />);
		expect(screen.getByText('folder')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'child' })).toBeInTheDocument();
		expect(screen.getByText(/FOLDER - 3 ITENS/)).toBeInTheDocument();
		fireEvent.click(screen.getByRole('button', { name: 'child' }));
		expect(mockOpenMediaItem).toHaveBeenCalledWith(expect.objectContaining({ id: 3, name: 'child' }));
		expect(handleSelectItem).toHaveBeenCalledWith(3);
		fireEvent.click(screen.getByRole('button', { name: 'star-child' }));
		expect(handleStarredItem).toHaveBeenCalledWith(3);

		mockUseFile.mockReturnValue({
			status: 'success',
			handleSelectItem: jest.fn(),
			handleStarredItem: jest.fn(),
			selectedItem: { id: 9, type: 2, name: 'report.pdf', format: '.pdf', size: 500 },
			files: [],
		});
		render(<FileContent />);
		expect(screen.getByText('viewer:report.pdf')).toBeInTheDocument();
	});

	it('does not reselect files handled by the shared media opener', () => {
		const handleSelectItem = jest.fn();
		mockOpenMediaItem.mockReturnValue(true);
		mockUseFile.mockReturnValue({
			status: 'success',
			handleSelectItem,
			handleStarredItem: jest.fn(),
			selectedItem: null,
			fileListFilter: 'all',
			files: [
				{ id: 1, type: 2, name: 'movie.mp4', format: '.mp4', size: 1024, starred: false },
			],
		});

		render(<FileContent />);
		fireEvent.click(screen.getByRole('button', { name: 'movie.mp4' }));

		expect(mockOpenMediaItem).toHaveBeenCalledWith(expect.objectContaining({ id: 1, name: 'movie.mp4' }));
		expect(handleSelectItem).not.toHaveBeenCalled();
	});

	it('supports list view without duplicated heading', () => {
		mockUseFile.mockReturnValue({
			status: 'success',
			handleSelectItem: jest.fn(),
			handleStarredItem: jest.fn(),
			selectedItem: null,
			fileListFilter: 'all',
			files: [
				{ id: 1, type: 2, name: 'song', format: '.mp3', size: 1024, starred: false },
			],
		});

		render(<FileContent viewMode='list' showHeading={false} />);
		expect(screen.queryByText('FILES')).not.toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'song' })).toBeInTheDocument();
	});

	it('supports custom collection data and empty state messages', () => {
		mockUseFile.mockReturnValue({
			status: 'success',
			handleSelectItem: jest.fn(),
			handleStarredItem: jest.fn(),
			selectedItem: null,
			fileListFilter: 'starred',
			files: [],
		});

		const { rerender } = render(
			<FileContent
				title='Favorites scope'
				items={[
					{ id: 7, type: 2, name: 'notes.txt', format: '.txt', size: 12, starred: true },
				]}
				emptyStateMessage='EMPTY_FAVORITES'
			/>,
		);

		expect(screen.getByText('Favorites scope')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'notes.txt' })).toBeInTheDocument();

		rerender(<FileContent title='Favorites scope' items={[]} emptyStateMessage='EMPTY_FAVORITES' />);
		expect(screen.getByText('EMPTY_FAVORITES')).toBeInTheDocument();
	});
});
