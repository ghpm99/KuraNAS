import { act, renderHook } from '@testing-library/react';
import { useDebouncedNameQuery } from './useDebouncedNameQuery';

describe('useDebouncedNameQuery', () => {
    beforeEach(() => {
        jest.useFakeTimers();
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    it('commits the typed text only after the debounce delay', () => {
        const commit = jest.fn();
        const { result } = renderHook(() => useDebouncedNameQuery('', commit));

        act(() => result.current.setTypedNameQuery('bea'));
        act(() => result.current.setTypedNameQuery('beach'));
        act(() => {
            jest.advanceTimersByTime(299);
        });
        expect(commit).not.toHaveBeenCalled();

        act(() => {
            jest.advanceTimersByTime(1);
        });
        expect(commit).toHaveBeenCalledTimes(1);
        expect(commit).toHaveBeenCalledWith('beach');
    });

    it('follows external changes of the committed value without committing back', () => {
        const commit = jest.fn();
        const { result, rerender } = renderHook(
            ({ committed }) => useDebouncedNameQuery(committed, commit),
            { initialProps: { committed: 'beach' } }
        );
        expect(result.current.typedNameQuery).toBe('beach');

        rerender({ committed: '' });
        act(() => {
            jest.advanceTimersByTime(1000);
        });

        expect(result.current.typedNameQuery).toBe('');
        expect(commit).not.toHaveBeenCalled();
    });
});
