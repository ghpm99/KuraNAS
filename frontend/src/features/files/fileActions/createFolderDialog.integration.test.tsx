import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import FileProvider from '@/features/files/providers/fileProvider';
import { apiBase } from '@/service';
import useCreateFolderFlow from './useCreateFolderFlow';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn(), delete: jest.fn() },
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock; delete: jest.Mock };

const CreateFolderLauncher = () => {
    const { openCreateFolderDialog, createFolderDialog } = useCreateFolderFlow();
    return (
        <>
            <button onClick={openCreateFolderDialog}>launch-create-folder</button>
            {createFolderDialog}
        </>
    );
};

const renderLauncher = () =>
    render(
        <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
            <MemoryRouter initialEntries={['/files']}>
                <FileProvider>
                    <CreateFolderLauncher />
                </FileProvider>
            </MemoryRouter>
        </QueryClientProvider>
    );

describe('create folder dialog (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockResolvedValue({
            data: { items: [], pagination: { has_next: false, page: 1 } },
        });
        mockedApi.post.mockResolvedValue({ data: undefined });
    });

    it('posts the exact create-folder payload', async () => {
        renderLauncher();

        fireEvent.click(screen.getByText('launch-create-folder'));
        fireEvent.change(await screen.findByLabelText('NAME'), { target: { value: ' Photos ' } });
        fireEvent.click(screen.getAllByRole('button', { name: 'NEW_FOLDER' }).pop()!);

        await waitFor(() =>
            expect(mockedApi.post).toHaveBeenCalledWith('/files/folder', {
                name: 'Photos',
                parent_id: null,
            })
        );
    });

    it('shows the backend conflict message verbatim inside the dialog', async () => {
        mockedApi.post.mockRejectedValue({ response: { data: { error: 'Pasta já existe' } } });
        renderLauncher();

        fireEvent.click(screen.getByText('launch-create-folder'));
        fireEvent.change(await screen.findByLabelText('NAME'), { target: { value: 'Photos' } });
        fireEvent.click(screen.getAllByRole('button', { name: 'NEW_FOLDER' }).pop()!);

        expect(await screen.findByRole('alert')).toHaveTextContent('Pasta já existe');
        expect(screen.getByRole('dialog')).toBeInTheDocument();
    });

    it('falls back to the generic message when the backend sent none', async () => {
        mockedApi.post.mockRejectedValue(new Error('network'));
        renderLauncher();

        fireEvent.click(screen.getByText('launch-create-folder'));
        fireEvent.change(await screen.findByLabelText('NAME'), { target: { value: 'Photos' } });
        fireEvent.click(screen.getAllByRole('button', { name: 'NEW_FOLDER' }).pop()!);

        expect(await screen.findByRole('alert')).toHaveTextContent('ERROR_CREATE_FOLDER_FAILED');
    });
});
