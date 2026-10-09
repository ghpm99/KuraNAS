import { act, fireEvent, render, screen } from '@testing-library/react';
import useInfiniteScrollSentinel from './useInfiniteScrollSentinel';

type ObserverCallback = (entries: Array<{ isIntersecting: boolean }>) => void;

const SentinelHarness = ({
    hasNextPage = true,
    isFetchingNextPage = false,
    fetchNextPage = () => undefined,
}: {
    hasNextPage?: boolean;
    isFetchingNextPage?: boolean;
    fetchNextPage?: () => void;
}) => {
    const { sentinelRef, loadMore } = useInfiniteScrollSentinel({
        hasNextPage,
        isFetchingNextPage,
        fetchNextPage,
    });
    return (
        <div>
            <div ref={sentinelRef} data-testid="sentinel" />
            <button onClick={loadMore}>load</button>
        </div>
    );
};

describe('useInfiniteScrollSentinel', () => {
    const originalObserver = globalThis.IntersectionObserver;

    afterEach(() => {
        globalThis.IntersectionObserver = originalObserver;
    });

    it('renders without IntersectionObserver and still loads through loadMore', () => {
        // @ts-expect-error simulating an environment without IntersectionObserver
        delete globalThis.IntersectionObserver;
        const fetchNextPage = jest.fn();
        render(<SentinelHarness fetchNextPage={fetchNextPage} />);

        fireEvent.click(screen.getByText('load'));

        expect(fetchNextPage).toHaveBeenCalledTimes(1);
    });

    it('fetches the next page when the sentinel becomes visible', () => {
        let triggerObserver: ObserverCallback = () => undefined;
        const disconnect = jest.fn();
        globalThis.IntersectionObserver = jest.fn((callback: ObserverCallback) => {
            triggerObserver = callback;
            return { observe: jest.fn(), disconnect, unobserve: jest.fn() };
        }) as unknown as typeof IntersectionObserver;
        const fetchNextPage = jest.fn();
        const { unmount } = render(<SentinelHarness fetchNextPage={fetchNextPage} />);

        act(() => triggerObserver([{ isIntersecting: false }]));
        expect(fetchNextPage).not.toHaveBeenCalled();

        act(() => triggerObserver([{ isIntersecting: true }]));
        expect(fetchNextPage).toHaveBeenCalledTimes(1);

        unmount();
        expect(disconnect).toHaveBeenCalled();
    });

    it('does not observe or fetch when there is no next page or a fetch is running', () => {
        const observerConstructor = jest.fn();
        globalThis.IntersectionObserver = observerConstructor as unknown as typeof IntersectionObserver;
        const fetchNextPage = jest.fn();

        const { rerender } = render(<SentinelHarness hasNextPage={false} fetchNextPage={fetchNextPage} />);
        fireEvent.click(screen.getByText('load'));
        rerender(<SentinelHarness isFetchingNextPage fetchNextPage={fetchNextPage} />);
        fireEvent.click(screen.getByText('load'));

        expect(observerConstructor).not.toHaveBeenCalled();
        expect(fetchNextPage).not.toHaveBeenCalled();
    });
});
