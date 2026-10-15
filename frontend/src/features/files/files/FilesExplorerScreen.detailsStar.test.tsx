import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import SettingsProvider from '@/components/providers/settingsProvider';
import { GlobalMusicProvider } from '@/features/music/providers/GlobalMusicProvider';
import FileProvider from '@/features/files/providers/fileProvider';
import FilesExplorerScreen from './FilesExplorerScreen';

const mockGetFilesTree = jest.fn();
const mockGetFileByPath = jest.fn();
const mockGetFileAncestors = jest.fn();
const mockSetFileStarred = jest.fn();

jest.mock('@/service/files', () => ({
    ...jest.requireActual('@/service/files'),
    getFilesTree: (...args: unknown[]) => mockGetFilesTree(...args),
    getFileByPath: (...args: unknown[]) => mockGetFileByPath(...args),
    getFileAncestors: (...args: unknown[]) => mockGetFileAncestors(...args),
    setFileStarred: (...args: unknown[]) => mockSetFileStarred(...args),
    getRecentAccessByFileId: jest.fn().mockResolvedValue([]),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const folderToInspect: FileData = {
    id: 20,
    name: 'Projects',
    path: '/Projects',
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
    directory_content_count: 0,
    starred: false,
};

const renderExplorer = () =>
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <SettingsProvider>
                <GlobalMusicProvider>
                    <MemoryRouter initialEntries={['/files']}>
                        <FileProvider>
                            <FilesExplorerScreen />
                        </FileProvider>
                    </MemoryRouter>
                </GlobalMusicProvider>
            </SettingsProvider>
        </QueryClientProvider>
    );

describe('Starring from the details panel opened by the explicit action', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        let isStarredOnServer = false;
        mockGetFilesTree.mockImplementation(async () => ({
            items: [{ ...folderToInspect, starred: isStarredOnServer }],
            pagination: { hasNext: false, hasPrevious: false, page: 1, pageSize: 200 },
        }));
        mockGetFileByPath.mockResolvedValue(null);
        mockGetFileAncestors.mockResolvedValue([]);
        mockSetFileStarred.mockImplementation(async (fileId: number, isStarred: boolean) => {
            isStarredOnServer = isStarred;
            return { id: fileId, starred: isStarred };
        });
    });

    it('swaps the star icon and label right after the click', async () => {
        renderExplorer();

        fireEvent.click(await screen.findByLabelText('FILES_ITEM_MENU'));
        fireEvent.click(await screen.findByText('FILES_DETAILS'));

        const detailsPanel = (await screen.findByText('FILE_DETAILS_TITLE')).closest('div')
            ?.parentElement as HTMLElement;
        fireEvent.click(within(detailsPanel).getByRole('button', { name: 'FILES_FAVORITE' }));

        await waitFor(() =>
            expect(
                within(detailsPanel).getByRole('button', { name: 'FILES_UNFAVORITE' })
            ).toHaveAttribute('aria-pressed', 'true')
        );
        expect(mockSetFileStarred).toHaveBeenCalledWith(20, true);
    });
});
