import { act, renderHook } from '@testing-library/react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { startImageClassificationBackfill } from '@/service/files';
import useImageClassificationBackfill from './useImageClassificationBackfill';

const mockEnqueueSnackbar = jest.fn();
const mockInvalidateQueries = jest.fn();

jest.mock('@tanstack/react-query', () => ({
    useQuery: jest.fn(),
    useMutation: jest.fn(),
    useQueryClient: jest.fn(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

jest.mock('notistack', () => ({
    useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

jest.mock('@/service/files', () => ({
    getPendingImageClassificationCount: jest.fn(),
    startImageClassificationBackfill: jest.fn(),
}));

const mockedUseQuery = useQuery as jest.Mock;
const mockedUseMutation = useMutation as jest.Mock;
const mockedUseQueryClient = useQueryClient as jest.Mock;
const mockedStart = startImageClassificationBackfill as jest.Mock;

describe('components/settings/useImageClassificationBackfill', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedUseQueryClient.mockReturnValue({ invalidateQueries: mockInvalidateQueries });
        mockedUseQuery.mockReturnValue({ data: 4, isLoading: false, isError: false });
        mockedUseMutation.mockImplementation(
            ({
                mutationFn,
                onSuccess,
                onError,
            }: {
                mutationFn: () => Promise<unknown>;
                onSuccess?: (result: unknown) => void;
                onError?: (error: unknown) => void;
            }) => ({
                mutate: async () => {
                    try {
                        const result = await mutationFn();
                        onSuccess?.(result);
                    } catch (error) {
                        onError?.(error);
                    }
                },
                isPending: false,
            })
        );
    });

    it('exposes the pending count from the query', () => {
        const { result } = renderHook(() => useImageClassificationBackfill());
        expect(result.current.pendingCount).toBe(4);
        expect(result.current.isLoading).toBe(false);
        expect(result.current.hasError).toBe(false);
    });

    it('defaults the pending count to zero before it loads', () => {
        mockedUseQuery.mockReturnValue({ data: undefined, isLoading: true, isError: false });
        const { result } = renderHook(() => useImageClassificationBackfill());
        expect(result.current.pendingCount).toBe(0);
        expect(result.current.isLoading).toBe(true);
    });

    it('starts the backfill and shows the success toast', async () => {
        mockedStart.mockResolvedValue(7);
        const { result } = renderHook(() => useImageClassificationBackfill());

        await act(async () => {
            result.current.startBackfill();
        });

        expect(mockedStart).toHaveBeenCalled();
        expect(mockInvalidateQueries).toHaveBeenCalledWith({
            queryKey: ['image-classification-pending-count'],
        });
        expect(mockEnqueueSnackbar).toHaveBeenCalledWith(
            'IMAGE_CLASSIFY_BACKFILL_TOAST_SUCCESS',
            { variant: 'success' }
        );
    });

    it('shows the backend message verbatim when starting fails', async () => {
        mockedStart.mockRejectedValue({ response: { data: { error: 'indisponível' } } });
        const { result } = renderHook(() => useImageClassificationBackfill());

        await act(async () => {
            result.current.startBackfill();
        });

        expect(mockEnqueueSnackbar).toHaveBeenCalledWith('indisponível', { variant: 'error' });
    });

    it('falls back to the generic error without a backend message', async () => {
        mockedStart.mockRejectedValue(new Error('network'));
        const { result } = renderHook(() => useImageClassificationBackfill());

        await act(async () => {
            result.current.startBackfill();
        });

        expect(mockEnqueueSnackbar).toHaveBeenCalledWith(
            'IMAGE_CLASSIFY_BACKFILL_TOAST_ERROR',
            { variant: 'error' }
        );
    });
});
