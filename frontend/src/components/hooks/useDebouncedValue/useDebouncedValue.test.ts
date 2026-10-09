import { act, renderHook } from '@testing-library/react';
import useDebouncedValue from './useDebouncedValue';

describe('useDebouncedValue', () => {
    beforeEach(() => jest.useFakeTimers());
    afterEach(() => jest.useRealTimers());

    it('starts with the initial value without any timer elapsed', () => {
        const { result } = renderHook(() => useDebouncedValue('first', 300));

        expect(result.current).toBe('first');
    });

    it('only publishes the latest value after the delay', () => {
        const { result, rerender } = renderHook(({ value }) => useDebouncedValue(value, 300), {
            initialProps: { value: 'a' },
        });

        rerender({ value: 'ab' });
        act(() => {
            jest.advanceTimersByTime(299);
        });
        expect(result.current).toBe('a');

        rerender({ value: 'abc' });
        act(() => {
            jest.advanceTimersByTime(299);
        });
        expect(result.current).toBe('a');

        act(() => {
            jest.advanceTimersByTime(1);
        });
        expect(result.current).toBe('abc');
    });
});
