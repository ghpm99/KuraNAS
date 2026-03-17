import { render, screen } from '@testing-library/react';
import FileDetails from './fileDetails';

const mockUseFile = jest.fn();

jest.mock('../providers/fileProvider/fileContext', () => ({
    __esModule: true,
    default: () => mockUseFile(),
}));

jest.mock('../i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (k: string) => k }),
}));

describe('fileDetails', () => {
    it('returns null for no item or directory', () => {
        mockUseFile.mockReturnValue({
            selectedItem: null,
            isLoadingAccessData: false,
            recentAccessFiles: [],
            handleSelectItem: jest.fn(),
        });
        const { container } = render(<FileDetails />);
        expect(container).toBeEmptyDOMElement();

        mockUseFile.mockReturnValue({
            selectedItem: { id: 1, type: 1 },
            isLoadingAccessData: false,
            recentAccessFiles: [],
            handleSelectItem: jest.fn(),
        });
        const d = render(<FileDetails />);
        expect(d.container).toBeEmptyDOMElement();
    });

    it('renders file details and recent access list', () => {
        mockUseFile.mockReturnValue({
            selectedItem: {
                id: 2,
                type: 2,
                format: '.mp3',
                size: 1024,
                created_at: '2026-01-01T00:00:00Z',
                updated_at: '2026-01-02T00:00:00Z',
                path: '/music/a.mp3',
            },
            isLoadingAccessData: false,
            recentAccessFiles: [
                {
                    id: 10,
                    ip_address: '127.0.0.1',
                    file_id: 2,
                    accessed_at: '2026-01-03T00:00:00Z',
                },
            ],
            handleSelectItem: jest.fn(),
        });

        render(<FileDetails />);
        expect(screen.getByText('FILE_DETAILS_TITLE')).toBeInTheDocument();
        expect(screen.getByText('/music/a.mp3')).toBeInTheDocument();
        expect(screen.getByText('127.0.0.1')).toBeInTheDocument();
    });

    it('renders loading spinner for recent activity', () => {
        mockUseFile.mockReturnValue({
            selectedItem: {
                id: 2,
                type: 2,
                format: '.mp3',
                size: 1024,
                created_at: '2026-01-01T00:00:00Z',
                updated_at: '2026-01-02T00:00:00Z',
                path: '/music/a.mp3',
            },
            isLoadingAccessData: true,
            recentAccessFiles: [],
            handleSelectItem: jest.fn(),
        });
        render(<FileDetails />);
        expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });
});
