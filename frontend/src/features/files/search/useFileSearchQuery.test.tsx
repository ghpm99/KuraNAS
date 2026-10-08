import type { ReactNode } from 'react';
import { act, renderHook } from '@testing-library/react';
import { MemoryRouter, useLocation } from 'react-router-dom';
import useFileSearchQuery from './useFileSearchQuery';

const renderQueryHook = (initialUrl: string) =>
    renderHook(
        () => ({ search: useFileSearchQuery(), location: useLocation() }),
        {
            wrapper: ({ children }: { children: ReactNode }) => (
                <MemoryRouter initialEntries={[initialUrl]}>{children}</MemoryRouter>
            ),
        }
    );

describe('useFileSearchQuery', () => {
    beforeEach(() => jest.useFakeTimers({ doNotFake: ['queueMicrotask'] }));
    afterEach(() => jest.useRealTimers());

    it('starts empty and inactive without a q param', () => {
        const { result } = renderQueryHook('/files');

        expect(result.current.search.inputValue).toBe('');
        expect(result.current.search.activeQuery).toBe('');
    });

    it('starts from the q param of a shared url', () => {
        const { result } = renderQueryHook('/files?q=relatorio');

        expect(result.current.search.inputValue).toBe('relatorio');
        expect(result.current.search.activeQuery).toBe('relatorio');
    });

    it('debounces typing by 300ms before writing q to the url', async () => {
        const { result } = renderQueryHook('/files');

        act(() => result.current.search.setInputValue('rel'));
        act(() => {
            jest.advanceTimersByTime(299);
        });
        expect(result.current.location.search).toBe('');

        await act(async () => {
            jest.advanceTimersByTime(1);
        });
        expect(result.current.location.search).toBe('?q=rel');
        expect(result.current.search.activeQuery).toBe('rel');
    });

    it('keeps what the user kept typing while an earlier value is being written', async () => {
        const { result } = renderQueryHook('/files');

        act(() => result.current.search.setInputValue('rel'));
        await act(async () => {
            jest.advanceTimersByTime(300);
        });
        act(() => result.current.search.setInputValue('relat'));

        expect(result.current.search.inputValue).toBe('relat');
        await act(async () => {
            jest.advanceTimersByTime(300);
        });
        expect(result.current.location.search).toBe('?q=relat');
    });

    it('does not activate the search below the minimum length', async () => {
        const { result } = renderQueryHook('/files');

        act(() => result.current.search.setInputValue('a'));
        await act(async () => {
            jest.advanceTimersByTime(300);
        });

        expect(result.current.location.search).toBe('?q=a');
        expect(result.current.search.activeQuery).toBe('');
    });

    it('preserves other query params when writing q', async () => {
        const { result } = renderQueryHook('/files?view=list');

        act(() => result.current.search.setInputValue('foto'));
        await act(async () => {
            jest.advanceTimersByTime(300);
        });

        expect(result.current.location.search).toBe('?view=list&q=foto');
    });

    it('clears the field and the url immediately', () => {
        const { result } = renderQueryHook('/files?q=relatorio');

        act(() => result.current.search.clearQuery());

        expect(result.current.search.inputValue).toBe('');
        expect(result.current.location.search).toBe('');
        expect(result.current.search.activeQuery).toBe('');
    });
});
