import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import FileProvider from '@/features/files/providers/fileProvider';
import { createTestFile } from '@/features/files/selection/testFileFactory';
import { apiBase } from '@/service';
import useFileActionFlow from './useFileActionFlow';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn(), delete: jest.fn() },
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock; delete: jest.Mock };

const documentFile = createTestFile(42, { name: 'contract.pdf' });

const DeleteLauncher = () => {
    const { startAction, dialogs } = useFileActionFlow();
    return (
        <>
            <button onClick={() => startAction('delete', [documentFile])}>launch-delete</button>
            <button onClick={() => startAction('rename', [documentFile])}>launch-rename</button>
            {dialogs}
        </>
    );
};

const renderLauncher = () =>
    render(
        <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
            <MemoryRouter initialEntries={['/files']}>
                <FileProvider>
                    <DeleteLauncher />
                </FileProvider>
            </MemoryRouter>
        </QueryClientProvider>
    );

describe('delete and rename dialogs (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockResolvedValue({
            data: { items: [], pagination: { has_next: false, page: 1 } },
        });
        mockedApi.post.mockResolvedValue({ data: { path: '/library/new.pdf' } });
        mockedApi.delete.mockResolvedValue({ data: undefined });
    });

    it('sends a plain delete without the permanent flag by default', async () => {
        renderLauncher();

        fireEvent.click(screen.getByText('launch-delete'));
        fireEvent.click(await screen.findByRole('button', { name: 'DELETE' }));

        await waitFor(() =>
            expect(mockedApi.delete).toHaveBeenCalledWith('/files/path', { data: { id: 42 } })
        );
    });

    it('sends permanent=true when the permanent checkbox is checked', async () => {
        renderLauncher();

        fireEvent.click(screen.getByText('launch-delete'));
        fireEvent.click(await screen.findByRole('checkbox', { name: 'FILES_DELETE_PERMANENTLY' }));
        fireEvent.click(screen.getAllByRole('button', { name: 'FILES_DELETE_PERMANENTLY' })[0]!);

        await waitFor(() =>
            expect(mockedApi.delete).toHaveBeenCalledWith('/files/path', {
                data: { id: 42 },
                params: { permanent: true },
            })
        );
    });

    it('shows the backend rename conflict inside the dialog and posts the exact payload', async () => {
        mockedApi.post.mockRejectedValue({ response: { data: { error: 'Nome já em uso' } } });
        renderLauncher();

        fireEvent.click(screen.getByText('launch-rename'));
        fireEvent.change(await screen.findByLabelText('NAME'), { target: { value: 'taken.pdf' } });
        fireEvent.click(screen.getByRole('button', { name: 'RENAME' }));

        expect(await screen.findByRole('alert')).toHaveTextContent('Nome já em uso');
        expect(mockedApi.post).toHaveBeenCalledWith('/files/rename', {
            id: 42,
            new_name: 'taken.pdf',
        });
        expect(screen.getByRole('dialog')).toBeInTheDocument();
    });
});
