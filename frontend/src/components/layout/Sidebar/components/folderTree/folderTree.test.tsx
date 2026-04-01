import { fireEvent, render, screen } from '@testing-library/react';
import FolderTree from './index';

const mockUseFile = jest.fn();

jest.mock('@/features/files/providers/fileProvider/fileContext', () => ({
    __esModule: true,
    default: () => mockUseFile(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (k: string) => k }),
}));

jest.mock('./components/folderItem', () => ({ children, label, onClick }: any) => (
    <div>
        <button onClick={onClick}>{label}</button>
        {children}
    </div>
));

describe('folderTree', () => {
    it('renders loading and error states', () => {
        mockUseFile.mockReturnValue({
            status: 'loading',
            files: [],
            handleSelectItem: jest.fn(),
        });
        render(<FolderTree />);
        expect(screen.getByRole('progressbar')).toBeInTheDocument();

        mockUseFile.mockReturnValue({
            status: 'error',
            files: [],
            handleSelectItem: jest.fn(),
        });
        render(<FolderTree />);
        expect(screen.getByText('ERROR_LOADING_FILES')).toBeInTheDocument();
    });

    it('renders empty list and files tree', () => {
        const handleSelectItem = jest.fn();
        mockUseFile.mockReturnValue({
            status: 'success',
            files: [],
            handleSelectItem,
            expandedItems: [],
            selectedItem: null,
        });
        render(<FolderTree />);
        expect(screen.getByText('EMPTY_FILE_LIST')).toBeInTheDocument();

        mockUseFile.mockReturnValue({
            status: 'success',
            files: [{ id: 1, type: 1, name: 'Folder', file_children: [] }],
            handleSelectItem,
            expandedItems: [1],
            selectedItem: { id: 1 },
        });
        render(<FolderTree />);
        fireEvent.click(screen.getByText('Folder'));
        expect(handleSelectItem).toHaveBeenCalledWith(expect.objectContaining({ id: 1 }));
        fireEvent.click(screen.getAllByText('FILES')[1] as HTMLElement);
        expect(handleSelectItem).toHaveBeenCalledWith(null);
    });
});
