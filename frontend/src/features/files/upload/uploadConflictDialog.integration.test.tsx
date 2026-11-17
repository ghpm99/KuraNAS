import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { apiBase } from '@/service';
import UploadQueueProvider from './uploadQueueProvider';
import useUploadQueueContext from './uploadQueueContext';

jest.mock('@/service', () => ({
    apiBase: {
        post: jest.fn(),
        get: jest.fn(),
        delete: jest.fn(),
    },
}));

const mockedApi = apiBase as unknown as { post: jest.Mock };

const Enqueuer = () => {
    const { enqueue } = useUploadQueueContext();
    return (
        <button onClick={() => enqueue([{ file: new File(['conteudo'], 'foto.jpg') }], 4)}>
            enqueue
        </button>
    );
};

const renderQueue = () =>
    render(
        <QueryClientProvider client={new QueryClient()}>
            <UploadQueueProvider>
                <Enqueuer />
            </UploadQueueProvider>
        </QueryClientProvider>
    );

const onConflictSentInCall = (callIndex: number): FormDataEntryValue | null =>
    (mockedApi.post.mock.calls[callIndex]![1] as FormData).get('on_conflict');

describe('upload conflict dialog (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        window.localStorage.clear();
    });

    it('asks on a name conflict and resends with the chosen policy', async () => {
        mockedApi.post
            .mockResolvedValueOnce({
                data: {
                    files: [{ name: 'foto.jpg', status: 'conflict', error: 'Já existe' }],
                },
            })
            .mockResolvedValueOnce({
                data: { files: [{ name: 'foto.jpg', status: 'replaced' }] },
            });
        renderQueue();

        fireEvent.click(screen.getByRole('button', { name: 'enqueue' }));

        const dialog = await screen.findByRole('dialog');
        expect(mockedApi.post).toHaveBeenCalledTimes(1);
        expect(mockedApi.post.mock.calls[0]![0]).toBe('/files/upload');
        expect(onConflictSentInCall(0)).toBe('fail');

        fireEvent.click(
            within(dialog).getByRole('button', { name: 'FILES_UPLOAD_ON_CONFLICT_REPLACE' })
        );

        await waitFor(() => expect(mockedApi.post).toHaveBeenCalledTimes(2));
        expect(mockedApi.post.mock.calls[1]![0]).toBe('/files/upload');
        expect(onConflictSentInCall(1)).toBe('replace');
        await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
        expect(screen.getByText('FILES_UPLOAD_STATUS_REPLACED')).toBeInTheDocument();
    });
});
