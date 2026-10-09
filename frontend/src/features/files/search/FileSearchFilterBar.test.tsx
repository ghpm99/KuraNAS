import { fireEvent, render, screen } from '@testing-library/react';
import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import FileSearchFilterBar from './FileSearchFilterBar';
import { emptyFileSearchFilters } from './fileSearchFilters';

const mockMatchMedia = (matches: boolean) => {
    Object.defineProperty(window, 'matchMedia', {
        configurable: true,
        writable: true,
        value: (query: string) => ({
            matches,
            media: query,
            onchange: null,
            addListener: jest.fn(),
            removeListener: jest.fn(),
            addEventListener: jest.fn(),
            removeEventListener: jest.fn(),
            dispatchEvent: jest.fn(),
        }),
    });
};

describe('FileSearchFilterBar', () => {
    afterEach(() => {
        Reflect.deleteProperty(window, 'matchMedia');
    });

    it('renders without props and without a backend', () => {
        expectRendersWithoutBackend(<FileSearchFilterBar />);
    });

    it('shows the inline controls above the phone breakpoint', () => {
        mockMatchMedia(false);
        render(<FileSearchFilterBar />);

        expect(screen.getByRole('button', { name: 'FILES_SEARCH_FILTER_STARRED' })).toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'FILES_SEARCH_FILTERS_BUTTON' })).toBeNull();
        expect(screen.queryByRole('button', { name: 'FILES_SEARCH_FILTERS_CLEAR' })).toBeNull();
    });

    it('offers to clear the filters when some are active', () => {
        mockMatchMedia(false);
        const onReset = jest.fn();
        render(
            <FileSearchFilterBar
                filters={{ ...emptyFileSearchFilters, onlyStarred: true }}
                onReset={onReset}
            />
        );

        fireEvent.click(screen.getByRole('button', { name: 'FILES_SEARCH_FILTERS_CLEAR' }));

        expect(onReset).toHaveBeenCalledTimes(1);
    });

    it('collapses into a Filters button with a drawer below the phone breakpoint', () => {
        mockMatchMedia(true);
        const onChange = jest.fn();
        render(
            <FileSearchFilterBar
                filters={{ ...emptyFileSearchFilters, onlyStarred: true }}
                onChange={onChange}
            />
        );

        expect(screen.queryByRole('button', { name: 'FILES_SEARCH_FILTER_STARRED' })).toBeNull();
        expect(screen.getByText('1')).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'FILES_SEARCH_FILTERS_BUTTON' }));
        expect(screen.getByText('FILES_SEARCH_FILTERS_TITLE')).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'FILES_SEARCH_FILTER_STARRED' }));
        expect(onChange).toHaveBeenCalledWith({ ...emptyFileSearchFilters, onlyStarred: false });

        fireEvent.click(screen.getByRole('button', { name: 'FILES_SEARCH_FILTERS_CLOSE' }));
    });
});
