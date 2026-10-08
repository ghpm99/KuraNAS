import { fireEvent, render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter, useLocation } from 'react-router-dom';
import FilesExplorerScreen from './FilesExplorerScreen';

const mockUseFile = jest.fn();

jest.mock('@/features/files/providers/fileProvider/fileContext', () => ({
    __esModule: true,
    default: () => mockUseFile(),
}));

jest.mock('@/features/files/providers/fileProvider/useFileAncestors', () => ({
    __esModule: true,
    default: () => ({ data: undefined }),
}));

jest.mock('@/components/actionBar', () => () => <div>ActionBarMock</div>);
jest.mock(
    '@/features/files/fileContent',
    () =>
        ({ onGoToParent, onFocusSearch }: any) => (
            <div>
                <button onClick={onGoToParent}>go-parent</button>
                <button onClick={onFocusSearch}>focus-search</button>
            </div>
        )
);
jest.mock('@/features/files/fileDetails', () => () => <div>FileDetailsMock</div>);
jest.mock('@/components/layout/Sidebar/components/folderTree', () => () => <div>TreeMock</div>);
jest.mock('@/components/tabs', () => () => <div>TabsMock</div>);
jest.mock('@/features/files/findByDiskPath/findByDiskPathDialog', () => () => null);

const nestedFolder = {
    id: 7,
    name: 'docs',
    path: '/library/docs',
    parent_path: '/library',
    type: 1,
    file_children: [],
};

const LocationProbe = () => <span data-testid="location">{useLocation().pathname}</span>;

const renderScreen = (selectedItem: unknown) => {
    mockUseFile.mockReturnValue({
        files: [],
        selectedItem,
        handleSelectItem: jest.fn(),
        fileListFilter: 'all',
        filesSort: { key: 'name', order: 'asc' },
        setFilesSort: jest.fn(),
    });
    render(
        <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
            <MemoryRouter initialEntries={['/files/library/docs']}>
                <FilesExplorerScreen />
                <LocationProbe />
            </MemoryRouter>
        </QueryClientProvider>
    );
};

describe('FilesExplorerScreen shortcut wiring', () => {
    it('navigates to the parent folder of the open folder', () => {
        renderScreen(nestedFolder);

        fireEvent.click(screen.getByText('go-parent'));

        expect(screen.getByTestId('location')).toHaveTextContent('/files/library');
    });

    it('navigates to the files root from a top-level folder', () => {
        renderScreen({ ...nestedFolder, parent_path: '/' });

        fireEvent.click(screen.getByText('go-parent'));

        expect(screen.getByTestId('location')).toHaveTextContent(/^\/files\/?$/);
    });

    it('offers no parent navigation at the root', () => {
        renderScreen(null);

        expect(screen.getByText('go-parent')).toBeInTheDocument();
        fireEvent.click(screen.getByText('go-parent'));
        expect(screen.getByTestId('location')).toHaveTextContent('/files/library/docs');
    });

    it('focuses the folder search field', () => {
        renderScreen(nestedFolder);

        fireEvent.click(screen.getByText('focus-search'));

        expect(screen.getByPlaceholderText('FILES_SEARCH_PLACEHOLDER')).toHaveFocus();
    });
});
