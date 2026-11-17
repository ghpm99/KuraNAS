import { act, fireEvent, render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { transferFile } from './transferFile';
import UploadQueueProvider from './uploadQueueProvider';
import useUploadQueueContext from './uploadQueueContext';

jest.mock('./transferFile', () => ({
    transferFile: jest.fn().mockResolvedValue({ status: 'uploaded' }),
}));

const Enqueuer = () => {
    const { enqueue } = useUploadQueueContext();
    return (
        <button
            onClick={() =>
                enqueue([{ file: new File(['x'], 'a.txt') }, { file: new File(['x'], 'b.txt') }], 3)
            }
        >
            enqueue
        </button>
    );
};

describe('UploadQueueProvider', () => {
    it('renders its children without any service mock behavior and without the panel', () => {
        const queryClient = new QueryClient();
        render(
            <QueryClientProvider client={queryClient}>
                <UploadQueueProvider>
                    <span>child</span>
                </UploadQueueProvider>
            </QueryClientProvider>
        );

        expect(screen.getByText('child')).toBeInTheDocument();
        expect(screen.queryByRole('complementary')).not.toBeInTheDocument();
    });

    it('shows the panel after enqueueing and invalidates the listing once, throttled', async () => {
        jest.useFakeTimers();
        const queryClient = new QueryClient();
        const invalidateSpy = jest.spyOn(queryClient, 'invalidateQueries').mockResolvedValue();
        render(
            <QueryClientProvider client={queryClient}>
                <UploadQueueProvider>
                    <Enqueuer />
                </UploadQueueProvider>
            </QueryClientProvider>
        );

        await act(async () => {
            fireEvent.click(screen.getByText('enqueue'));
        });

        expect(screen.getByRole('complementary')).toBeInTheDocument();
        expect(invalidateSpy).not.toHaveBeenCalled();

        act(() => {
            jest.advanceTimersByTime(2000);
        });

        expect(invalidateSpy).toHaveBeenCalledTimes(3);
        jest.useRealTimers();
    });

    it('asks for confirmation to leave only while a transfer is active', async () => {
        let finishTransfer: () => void = () => undefined;
        jest.mocked(transferFile).mockImplementationOnce(
            () =>
                new Promise((resolve) => {
                    finishTransfer = () => resolve({ status: 'uploaded' });
                })
        );
        render(
            <QueryClientProvider client={new QueryClient()}>
                <UploadQueueProvider>
                    <Enqueuer />
                </UploadQueueProvider>
            </QueryClientProvider>
        );
        const leavePage = () => {
            const event = new Event('beforeunload', { cancelable: true });
            window.dispatchEvent(event);
            return event.defaultPrevented;
        };
        expect(leavePage()).toBe(false);

        await act(async () => {
            fireEvent.click(screen.getByText('enqueue'));
        });
        expect(leavePage()).toBe(true);

        await act(async () => finishTransfer());
        expect(leavePage()).toBe(false);
    });
});
