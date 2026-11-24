import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import DocumentsScreen from './documentsScreenInFileProvider';

const mockGet = jest.fn();
const mockPost = jest.fn();
const mockDelete = jest.fn();
const mockTriggerBrowserDownload = jest.fn();

jest.mock('@/service', () => ({
    apiBase: {
        get: (...args: unknown[]) => mockGet(...args),
        post: (...args: unknown[]) => mockPost(...args),
        delete: (...args: unknown[]) => mockDelete(...args),
    },
}));

jest.mock('@/service/browserDownload', () => ({
    triggerBrowserDownload: (...args: unknown[]) => mockTriggerBrowserDownload(...args),
}));

jest.mock('@/components/folderPicker/folderPicker', () => ({
    __esModule: true,
    default: ({
        open,
        onSelect,
    }: {
        open: boolean;
        onSelect: (destination: { folderId: number }) => void;
    }) => (open ? <button onClick={() => onSelect({ folderId: 99 })}>CONFIRM_PICKER</button> : null),
}));

const pagination = { page: 1, page_size: 50, has_next: false, has_prev: false };

const buildListItem = (id: number, name: string, starred = false) => ({
    id,
    name,
    format: '.pdf',
    parent_path: '/Casa',
    size: 1024,
    modified_at: '2026-01-01T00:00:00Z',
    created_at: '2026-01-01T00:00:00Z',
    starred,
    index_state: 'indexed',
});

const searchResult = {
    file_id: 9,
    name: 'contrato.pdf',
    path: '/Casa/contrato.pdf',
    parent_path: '/Casa',
    format: '.pdf',
    size: 2048,
    updated_at: '2026-03-01T10:00:00Z',
    snippet: 'cláusula',
    page: 4,
};

const requestsTo = (url: string) =>
    mockGet.mock.calls.filter(([requestedUrl]) => requestedUrl === url);

const routeRequests = (listedItems = [buildListItem(1, 'escritura.pdf'), buildListItem(2, 'iptu.pdf')]) => {
    mockGet.mockImplementation((url: string) => {
        if (url === '/documents/count') {
            return Promise.resolve({ data: { total: listedItems.length } });
        }
        if (url === '/documents') {
            return Promise.resolve({ data: { items: listedItems, pagination } });
        }
        if (url === '/documents/search') {
            return Promise.resolve({ data: { items: [searchResult], pagination } });
        }
        return Promise.reject(new Error('no backend'));
    });
};

const openItemMenu = async (itemIndex = 0) => {
    const menuButtons = await screen.findAllByRole('button', { name: 'FILES_ITEM_MENU' });
    fireEvent.click(menuButtons[itemIndex]!);
};

const selectItem = async (itemIndex: number) => {
    const checkboxes = await screen.findAllByRole('checkbox', { name: 'FILES_SELECT_ITEM' });
    fireEvent.click(checkboxes[itemIndex]!);
};

const expectListRefetchedAfter = async (action: () => Promise<void> | void) => {
    await screen.findAllByText('escritura.pdf');
    const listRequestsBefore = requestsTo('/documents').length;
    const countRequestsBefore = requestsTo('/documents/count').length;

    await action();

    await waitFor(() => {
        expect(requestsTo('/documents').length).toBeGreaterThan(listRequestsBefore);
        expect(requestsTo('/documents/count').length).toBeGreaterThan(countRequestsBefore);
    });
};

describe('DocumentsScreen item menu and batch (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        window.localStorage.clear();
        routeRequests();
        mockPost.mockResolvedValue({ data: { path: '/Casa/novo.pdf', updated_ids: [1], not_found_ids: [] } });
        mockDelete.mockResolvedValue({ data: undefined });
    });

    it('renders the controls without any backend', async () => {
        mockGet.mockRejectedValue(new Error('no backend'));

        expect(() => renderWithoutBackend(<DocumentsScreen />)).not.toThrow();

        expect(await screen.findByRole('alert')).toBeInTheDocument();
        expect(screen.queryByRole('toolbar')).toBeNull();
    });

    it('offers download, rename, move, favorite, details and delete in the item menu', async () => {
        renderWithoutBackend(<DocumentsScreen />);

        await openItemMenu();

        const menu = await screen.findByRole('menu');
        ['DOWNLOAD', 'RENAME', 'MOVE', 'FILES_FAVORITE', 'FILES_DETAILS', 'DELETE'].forEach(
            (label) => expect(within(menu).getByRole('menuitem', { name: label })).toBeInTheDocument()
        );
    });

    it('opens the same menu with a right click', async () => {
        renderWithoutBackend(<DocumentsScreen />);

        const row = (await screen.findAllByText('escritura.pdf'))[0]!.closest('li')!;
        fireEvent.contextMenu(row);

        expect(await screen.findByRole('menu')).toBeInTheDocument();
    });

    it('posts the exact rename request and refetches the list and count', async () => {
        renderWithoutBackend(<DocumentsScreen />);

        await expectListRefetchedAfter(async () => {
            await openItemMenu();
            fireEvent.click(await screen.findByRole('menuitem', { name: 'RENAME' }));
            fireEvent.change(await screen.findByLabelText('NAME'), {
                target: { value: 'novo.pdf' },
            });
            fireEvent.click(screen.getByRole('button', { name: 'RENAME' }));
            await waitFor(() =>
                expect(mockPost).toHaveBeenCalledWith('/files/rename', {
                    id: 1,
                    new_name: 'novo.pdf',
                })
            );
        });
    });

    it('posts the exact move request and refetches the list and count', async () => {
        renderWithoutBackend(<DocumentsScreen />);

        await expectListRefetchedAfter(async () => {
            await openItemMenu();
            fireEvent.click(await screen.findByRole('menuitem', { name: 'MOVE' }));
            fireEvent.click(await screen.findByText('CONFIRM_PICKER'));
            await waitFor(() =>
                expect(mockPost).toHaveBeenCalledWith('/files/move', {
                    source_id: 1,
                    destination_folder_id: 99,
                    destination_path: '',
                })
            );
        });
    });

    it('posts the exact delete request and refetches the list and count', async () => {
        renderWithoutBackend(<DocumentsScreen />);

        await expectListRefetchedAfter(async () => {
            await openItemMenu();
            fireEvent.click(await screen.findByRole('menuitem', { name: 'DELETE' }));
            fireEvent.click(await screen.findByRole('button', { name: 'DELETE' }));
            await waitFor(() =>
                expect(mockDelete).toHaveBeenCalledWith('/files/path', { data: { id: 1 } })
            );
        });
    });

    it('favorites from the star button with the exact request and refetches', async () => {
        renderWithoutBackend(<DocumentsScreen />);

        await expectListRefetchedAfter(async () => {
            const starButtons = await screen.findAllByRole('button', {
                name: 'FILES_FAVORITE_ITEM',
            });
            fireEvent.click(starButtons[0]!);
            await waitFor(() =>
                expect(mockPost).toHaveBeenCalledWith('/files/starred', {
                    ids: [1],
                    starred: true,
                })
            );
        });
    });

    it('unfavorites a starred document from the menu', async () => {
        routeRequests([buildListItem(1, 'escritura.pdf', true)]);
        renderWithoutBackend(<DocumentsScreen />);

        await openItemMenu();
        fireEvent.click(await screen.findByRole('menuitem', { name: 'FILES_UNFAVORITE' }));

        await waitFor(() =>
            expect(mockPost).toHaveBeenCalledWith('/files/starred', { ids: [1], starred: false })
        );
    });

    it('downloads from the menu through the file endpoint', async () => {
        renderWithoutBackend(<DocumentsScreen />);

        await openItemMenu();
        fireEvent.click(await screen.findByRole('menuitem', { name: 'DOWNLOAD' }));

        expect(mockTriggerBrowserDownload).toHaveBeenCalledWith(
            expect.stringContaining('/files/download/1'),
            'escritura.pdf'
        );
    });

    it('opens the details drawer from the menu', async () => {
        renderWithoutBackend(<DocumentsScreen />);

        await openItemMenu();
        fireEvent.click(await screen.findByRole('menuitem', { name: 'FILES_DETAILS' }));

        expect(await screen.findByText('FILE_DETAILS_TITLE')).toBeInTheDocument();
    });

    it('selects documents, shows the batch bar and favorites them together', async () => {
        renderWithoutBackend(<DocumentsScreen />);

        await selectItem(0);
        await selectItem(1);

        const toolbar = screen.getByRole('toolbar');
        expect(within(toolbar).getByText('FILES_SELECTION_COUNT')).toBeInTheDocument();
        ['DOWNLOAD', 'MOVE', 'FILES_FAVORITE', 'DELETE'].forEach((label) =>
            expect(within(toolbar).getByRole('button', { name: label })).toBeInTheDocument()
        );
        expect(within(toolbar).queryByRole('button', { name: 'RENAME' })).toBeNull();

        mockPost.mockResolvedValue({ data: { updated_ids: [1, 2], not_found_ids: [] } });
        fireEvent.click(within(toolbar).getByRole('button', { name: 'FILES_FAVORITE' }));

        await waitFor(() =>
            expect(mockPost).toHaveBeenCalledWith('/files/starred', {
                ids: [1, 2],
                starred: true,
            })
        );
        await waitFor(() => expect(screen.queryByRole('toolbar')).toBeNull());
    });

    it('downloads the whole selection as a zip and deletes each selected document', async () => {
        renderWithoutBackend(<DocumentsScreen />);

        await selectItem(0);
        fireEvent.click(screen.getByRole('button', { name: 'FILES_SELECT_ALL' }));

        const toolbar = screen.getByRole('toolbar');
        expect(within(toolbar).getByText('FILES_SELECTION_COUNT')).toBeInTheDocument();
        fireEvent.click(within(toolbar).getByRole('button', { name: 'DOWNLOAD' }));
        expect(mockTriggerBrowserDownload).toHaveBeenCalledWith(
            expect.stringContaining('/files/download')
        );

        fireEvent.click(within(toolbar).getByRole('button', { name: 'DELETE' }));
        fireEvent.click(await screen.findByRole('button', { name: 'DELETE' }));

        await waitFor(() => expect(mockDelete).toHaveBeenCalledTimes(2));
        expect(mockDelete).toHaveBeenCalledWith('/files/path', { data: { id: 1 } });
        expect(mockDelete).toHaveBeenCalledWith('/files/path', { data: { id: 2 } });
    });

    it('clears the selection', async () => {
        renderWithoutBackend(<DocumentsScreen />);

        await selectItem(0);
        fireEvent.click(screen.getByRole('button', { name: 'FILES_CLEAR_SELECTION' }));

        expect(screen.queryByRole('toolbar')).toBeNull();
    });

    it('keeps the selection on reload and drops it when the sort changes', async () => {
        renderWithoutBackend(<DocumentsScreen />);

        await selectItem(0);
        fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_RELOAD' }));
        expect(screen.getByRole('toolbar')).toBeInTheDocument();

        fireEvent.mouseDown(screen.getByRole('combobox'));
        fireEvent.click(within(screen.getByRole('listbox')).getByText('DOCUMENTS_SORT_NAME'));

        await waitFor(() => expect(screen.queryByRole('toolbar')).toBeNull());
    });

    it('renames from a search result with the logical path of the result', async () => {
        renderWithoutBackend(<DocumentsScreen />, ['/documents?q=contrato']);

        await openItemMenu();
        fireEvent.click(await screen.findByRole('menuitem', { name: 'RENAME' }));
        fireEvent.change(await screen.findByLabelText('NAME'), { target: { value: 'novo.pdf' } });
        fireEvent.click(screen.getByRole('button', { name: 'RENAME' }));

        await waitFor(() =>
            expect(mockPost).toHaveBeenCalledWith('/files/rename', {
                id: 9,
                new_name: 'novo.pdf',
            })
        );
    });

    it('selects search results and shows the batch bar', async () => {
        renderWithoutBackend(<DocumentsScreen />, ['/documents?q=contrato']);

        await selectItem(0);

        expect(screen.getByRole('toolbar')).toBeInTheDocument();
    });
});
