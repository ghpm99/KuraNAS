import { QueryClient } from '@tanstack/react-query';
import { createThrottledListingInvalidation } from './invalidateListingThrottled';

describe('createThrottledListingInvalidation', () => {
    beforeEach(() => jest.useFakeTimers());
    afterEach(() => jest.useRealTimers());

    it('collapses a burst of requests into one invalidation of the listing queries', () => {
        const queryClient = new QueryClient();
        const invalidateSpy = jest.spyOn(queryClient, 'invalidateQueries').mockResolvedValue();
        const throttled = createThrottledListingInvalidation(queryClient, 1000);

        throttled.request();
        throttled.request();
        throttled.request();
        expect(invalidateSpy).not.toHaveBeenCalled();

        jest.advanceTimersByTime(1000);

        expect(invalidateSpy.mock.calls.map(([filters]) => filters)).toEqual([
            { queryKey: ['files'] },
            { queryKey: ['files-path'] },
            { queryKey: ['filesRecent'] },
        ]);

        throttled.request();
        jest.advanceTimersByTime(1000);
        expect(invalidateSpy).toHaveBeenCalledTimes(6);
    });

    it('cancel drops a pending invalidation', () => {
        const queryClient = new QueryClient();
        const invalidateSpy = jest.spyOn(queryClient, 'invalidateQueries').mockResolvedValue();
        const throttled = createThrottledListingInvalidation(queryClient, 1000);

        throttled.cancel();
        throttled.request();
        throttled.cancel();
        jest.advanceTimersByTime(2000);

        expect(invalidateSpy).not.toHaveBeenCalled();
    });
});
