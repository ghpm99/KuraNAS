import { fireEvent, screen, waitFor } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { UploadQueueContextProvider } from '@/features/files/upload/uploadQueueContext';
import type { UploadQueue } from '@/features/files/upload/uploadQueueTypes';
import DocumentsScreen from './documentsScreenInFileProvider';

const mockGet = jest.fn();
const mockPost = jest.fn();
const mockEnqueue = jest.fn();

jest.mock('@/service', () => ({
    apiBase: {
        get: (...args: unknown[]) => mockGet(...args),
        post: (...args: unknown[]) => mockPost(...args),
    },
}));

const pagination = { page: 1, page_size: 50, has_next: false, has_prev: false };

const requestsTo = (url: string) =>
    mockGet.mock.calls.filter(([requestedUrl]) => requestedUrl === url);

const routeRequests = (libraries: unknown = [{ category: 'documents', path: '/data/docs' }]) => {
    mockGet.mockImplementation((url: string) => {
        if (url === '/libraries') return Promise.resolve({ data: libraries });
        if (url === '/files/by-disk-path') return Promise.resolve({ data: { id: 10 } });
        if (url === '/files/tree') {
            return Promise.resolve({
                data: { items: [{ id: 11, name: 'Contratos', type: 1 }], pagination },
            });
        }
        if (url === '/documents/count') return Promise.resolve({ data: { total: 0 } });
        if (url === '/documents') return Promise.resolve({ data: { items: [], pagination } });
        return Promise.reject(new Error('no backend'));
    });
};

const renderScreen = (initialEntry = '/documents') =>
    renderWithoutBackend(
        <UploadQueueContextProvider value={{ enqueue: mockEnqueue } as unknown as UploadQueue}>
            <DocumentsScreen />
        </UploadQueueContextProvider>,
        [initialEntry]
    );

const pickFiles = (files: File[]) => {
    fireEvent.change(screen.getByTestId('documents-file-picker'), { target: { files } });
};

const confirmFolderName = async (folderName: string) => {
    fireEvent.click(await screen.findByRole('button', { name: 'NEW_FOLDER' }));
    fireEvent.change(await screen.findByLabelText('NAME'), { target: { value: folderName } });
    fireEvent.click(screen.getAllByRole('button', { name: 'NEW_FOLDER' }).pop()!);
};

describe('DocumentsScreen upload and new folder (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        window.localStorage.clear();
        routeRequests();
        mockPost.mockResolvedValue({ data: undefined });
    });

    it('renders the actions without any backend', async () => {
        mockGet.mockRejectedValue(new Error('no backend'));

        expect(() => renderScreen()).not.toThrow();

        expect(await screen.findByRole('button', { name: 'UPLOAD_FILE' })).toBeDisabled();
        expect(screen.getByRole('button', { name: 'NEW_FOLDER' })).toBeDisabled();
    });

    it('enqueues the picked files into the documents root when no folder is selected', async () => {
        renderScreen();
        await waitFor(() =>
            expect(screen.getByRole('button', { name: 'UPLOAD_FILE' })).toBeEnabled()
        );
        const pickedFile = new File(['pdf'], 'laudo.pdf', { type: 'application/pdf' });

        pickFiles([pickedFile]);

        expect(mockEnqueue).toHaveBeenCalledWith([{ file: pickedFile, relativePath: undefined }], 10);
    });

    it('enqueues into the folder selected in the URL', async () => {
        renderScreen('/documents?folder_id=11');
        await waitFor(() =>
            expect(screen.getByRole('button', { name: 'UPLOAD_FILE' })).toBeEnabled()
        );

        pickFiles([new File(['pdf'], 'laudo.pdf')]);

        expect(mockEnqueue).toHaveBeenCalledWith(expect.any(Array), 11);
    });

    it('posts the exact create-folder body under the documents root and refreshes the folder panel', async () => {
        renderScreen();
        await screen.findByText('Contratos');
        await waitFor(() => expect(screen.getByRole('button', { name: 'NEW_FOLDER' })).toBeEnabled());
        const treeRequestsBefore = requestsTo('/files/tree').length;

        await confirmFolderName('Escrituras');

        await waitFor(() =>
            expect(mockPost).toHaveBeenCalledWith('/files/folder', {
                name: 'Escrituras',
                parent_id: 10,
            })
        );
        await waitFor(() =>
            expect(requestsTo('/files/tree').length).toBeGreaterThan(treeRequestsBefore)
        );
    });

    it('creates the folder under the selected folder', async () => {
        renderScreen('/documents?folder_id=11');
        await waitFor(() => expect(screen.getByRole('button', { name: 'NEW_FOLDER' })).toBeEnabled());

        await confirmFolderName('Escrituras');

        await waitFor(() =>
            expect(mockPost).toHaveBeenCalledWith('/files/folder', {
                name: 'Escrituras',
                parent_id: 11,
            })
        );
    });

    it('shows the server error verbatim and keeps the dialog open', async () => {
        mockPost.mockRejectedValue({ response: { data: { error: 'Pasta já existe' } } });
        renderScreen();
        await waitFor(() => expect(screen.getByRole('button', { name: 'NEW_FOLDER' })).toBeEnabled());

        await confirmFolderName('Escrituras');

        expect(await screen.findByRole('alert')).toHaveTextContent('Pasta já existe');
        expect(screen.getByRole('dialog')).toBeInTheDocument();
    });

    it('disables both actions with an explanation when the library is not configured', async () => {
        routeRequests([]);

        renderScreen();

        await waitFor(() =>
            expect(screen.getByText('DOCUMENTS_FOLDERS_NOT_CONFIGURED')).toBeInTheDocument()
        );
        expect(screen.getByRole('button', { name: 'UPLOAD_FILE' })).toBeDisabled();
        expect(screen.getByRole('button', { name: 'NEW_FOLDER' })).toBeDisabled();
        fireEvent.mouseOver(screen.getByRole('button', { name: 'UPLOAD_FILE' }).parentElement!);
        expect(await screen.findByText('DOCUMENTS_ACTIONS_NOT_CONFIGURED')).toBeInTheDocument();
    });
});
