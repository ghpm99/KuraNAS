import { useState } from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter, useLocation } from 'react-router-dom';
import FileListFilterSync from '@/features/files/files/FileListFilterSync';
import { parseFileListFilter } from '@/features/files/files/fileListFilterFromSearch';
import type { FileListCategoryType } from '@/features/files/providers/fileProvider/fileContext';
import Tabs from './tabs';

const mockFileContext = {
    selectedItem: null as { id: number; type: number } | null,
    fileListFilter: 'all' as FileListCategoryType,
    setFileListFilter: (_filter: FileListCategoryType): void => undefined,
};

jest.mock('@/features/files/providers/fileProvider/fileContext', () => ({
    __esModule: true,
    default: () => mockFileContext,
}));

jest.mock('../i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const LocationProbe = () => {
    const location = useLocation();
    return <output aria-label="current-search">{location.search}</output>;
};

const FilesPageUnderTest = () => {
    const location = useLocation();
    const [fileListFilter, setFileListFilter] = useState<FileListCategoryType>('all');
    mockFileContext.fileListFilter = fileListFilter;
    mockFileContext.setFileListFilter = setFileListFilter;
    return (
        <>
            <FileListFilterSync filter={parseFileListFilter(location.search)} />
            <Tabs />
            <LocationProbe />
        </>
    );
};

const renderFilesPage = (initialEntry = '/files') =>
    render(
        <MemoryRouter initialEntries={[initialEntry]}>
            <FilesPageUnderTest />
        </MemoryRouter>
    );

const selectedTabName = () => screen.getByRole('tab', { selected: true }).textContent;

describe('components/tabs', () => {
    beforeEach(() => {
        mockFileContext.selectedItem = null;
    });

    it('switches to the clicked tab and keeps it selected', () => {
        renderFilesPage();

        fireEvent.click(screen.getByRole('tab', { name: 'RECENT_FILES' }));

        expect(selectedTabName()).toBe('RECENT_FILES');
        expect(screen.getByLabelText('current-search')).toHaveTextContent('?filter=recent');

        fireEvent.click(screen.getByRole('tab', { name: 'STARRED_FILES' }));

        expect(selectedTabName()).toBe('STARRED_FILES');
        expect(screen.getByLabelText('current-search')).toHaveTextContent('?filter=starred');
    });

    it('going back to all removes the filter from the url', () => {
        renderFilesPage('/files?filter=recent');
        expect(selectedTabName()).toBe('RECENT_FILES');

        fireEvent.click(screen.getByRole('tab', { name: 'ALL_FILES' }));

        expect(selectedTabName()).toBe('ALL_FILES');
        expect(screen.getByLabelText('current-search')).toBeEmptyDOMElement();
    });

    it('keeps other query parameters when switching tabs', () => {
        renderFilesPage('/files?sort=name');

        fireEvent.click(screen.getByRole('tab', { name: 'RECENT_FILES' }));

        expect(screen.getByLabelText('current-search')).toHaveTextContent(
            '?sort=name&filter=recent'
        );
    });

    it('returns null when selected item is a file', () => {
        mockFileContext.selectedItem = { id: 1, type: 2 };
        renderFilesPage();
        expect(screen.queryByRole('tab')).not.toBeInTheDocument();
    });
});
