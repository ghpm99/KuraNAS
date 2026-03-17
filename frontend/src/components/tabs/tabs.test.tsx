import { fireEvent, render, screen } from '@testing-library/react';
import Tabs from './tabs';

const mockUseFile = jest.fn();
const mockSetFileListFilter = jest.fn();

jest.mock('../providers/fileProvider/fileContext', () => ({
    __esModule: true,
    default: () => mockUseFile(),
}));

jest.mock('../i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (k: string) => k }),
}));

describe('components/tabs', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockUseFile.mockReturnValue({
            selectedItem: null,
            fileListFilter: 'all',
            setFileListFilter: mockSetFileListFilter,
        });
    });

    it('renders filters and changes selected tab', () => {
        render(<Tabs />);
        expect(screen.getByRole('tab', { name: 'ALL_FILES' })).toBeInTheDocument();
        fireEvent.click(screen.getByRole('tab', { name: 'STARRED_FILES' }));
        expect(mockSetFileListFilter).toHaveBeenCalledWith('starred');
    });

    it('returns null when selected item is a file', () => {
        mockUseFile.mockReturnValue({
            selectedItem: { id: 1, type: 2 },
            fileListFilter: 'all',
            setFileListFilter: mockSetFileListFilter,
        });
        const { container } = render(<Tabs />);
        expect(container).toBeEmptyDOMElement();
    });
});
