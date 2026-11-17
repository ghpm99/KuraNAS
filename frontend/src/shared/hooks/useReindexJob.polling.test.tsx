import { ReactNode } from 'react';
import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import useReindexJob from './useReindexJob';
import { requestReindex } from '@/service/indexing';
import { getJob } from '@/service/jobs';

jest.mock('@/service/indexing', () => ({ requestReindex: jest.fn() }));
jest.mock('@/service/jobs', () => ({ getJob: jest.fn() }));
jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockedRequestReindex = requestReindex as jest.Mock;
const mockedGetJob = getJob as jest.Mock;

const wrapper = ({ children }: { children: ReactNode }) => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return (
        <QueryClientProvider client={queryClient}>
            <SnackbarProvider>{children}</SnackbarProvider>
        </QueryClientProvider>
    );
};

const buildJob = (status: string, progress: number) => ({
    id: 3,
    type: 'reindex',
    status,
    progress: {
        total_steps: 10,
        completed_steps: 0,
        running_steps: 0,
        failed_steps: 0,
        skipped_steps: 0,
        canceled_steps: 0,
        progress,
    },
});

describe('useReindexJob polling', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedRequestReindex.mockResolvedValue({ jobId: 3, isAlreadyRunning: false });
    });

    it('calls onFinished once, only after the job completes', async () => {
        const onFinished = jest.fn();
        mockedGetJob
            .mockResolvedValueOnce(buildJob('running', 50))
            .mockResolvedValue(buildJob('completed', 100));
        const { result } = renderHook(() => useReindexJob({ onFinished }), { wrapper });

        await act(async () => {
            await result.current.start({ scope: 'folder', folderId: 8 });
        });

        expect(mockedRequestReindex).toHaveBeenCalledWith({ scope: 'folder', folderId: 8 });
        await waitFor(() => expect(result.current.progress).toBe(50));
        expect(result.current.isRunning).toBe(true);
        expect(onFinished).not.toHaveBeenCalled();

        await waitFor(() => expect(onFinished).toHaveBeenCalledTimes(1), { timeout: 5000 });
        expect(onFinished).toHaveBeenCalledWith(expect.objectContaining({ status: 'completed' }));
        expect(result.current.isRunning).toBe(false);
        expect(mockedGetJob.mock.calls.length).toBe(2);
    });

    it('follows the job already running reported by the server', async () => {
        mockedRequestReindex.mockResolvedValue({ jobId: 3, isAlreadyRunning: true });
        mockedGetJob.mockResolvedValue(buildJob('running', 10));
        const { result } = renderHook(() => useReindexJob(), { wrapper });

        await act(async () => {
            await result.current.start({ scope: 'all' });
        });

        await waitFor(() => expect(mockedGetJob).toHaveBeenCalledWith(3));
        expect(result.current.isRunning).toBe(true);
    });

    it('exposes the server message when the request fails', async () => {
        mockedRequestReindex.mockRejectedValue({
            response: { status: 404, data: { error: 'pasta não encontrada' } },
        });
        const { result } = renderHook(() => useReindexJob(), { wrapper });

        await act(async () => {
            await result.current.start({ scope: 'folder', folderId: 99 });
        });

        expect(result.current.errorMessage).toBe('pasta não encontrada');
        expect(result.current.isRunning).toBe(false);
        expect(mockedGetJob).not.toHaveBeenCalled();
    });

    it('calls onFinished with the failed job when it fails', async () => {
        const onFinished = jest.fn();
        mockedGetJob.mockResolvedValue({ ...buildJob('failed', 0), last_error: 'disco cheio' });
        const { result } = renderHook(() => useReindexJob({ onFinished }), { wrapper });

        await act(async () => {
            await result.current.start({ scope: 'all' });
        });

        await waitFor(() => expect(onFinished).toHaveBeenCalledTimes(1));
        expect(onFinished.mock.calls[0][0].status).toBe('failed');
    });
});
