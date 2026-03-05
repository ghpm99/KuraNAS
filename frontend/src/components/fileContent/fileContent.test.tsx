import { render, screen } from '@testing-library/react';
import { fireEvent } from '@testing-library/react';
import React from 'react';
import FileContent from './fileContent';

const mockUseFile = jest.fn();

jest.mock('../providers/fileProvider/fileContext', () => ({
	__esModule: true,
	default: () => mockUseFile(),
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
});
