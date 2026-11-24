import { fireEvent, render, screen } from '@testing-library/react';
import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import DocumentsFilterBar from './DocumentsFilterBar';
import { emptyDocumentsFilters } from './documentsFilters';

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

describe('DocumentsFilterBar', () => {
    afterEach(() => {
        Reflect.deleteProperty(window, 'matchMedia');
    });

    it('renders without props and without a backend', () => {
        expectRendersWithoutBackend(<DocumentsFilterBar />);
    });

    it('shows the inline controls above the phone breakpoint', () => {
        mockMatchMedia(false);
        render(<DocumentsFilterBar />);

        expect(
            screen.getByRole('button', { name: 'DOCUMENTS_FILTER_FAVORITES' })
        ).toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'DOCUMENTS_FILTERS_BUTTON' })).toBeNull();
        expect(screen.queryByRole('button', { name: 'DOCUMENTS_FILTERS_CLEAR' })).toBeNull();
    });

    it('offers to clear the filters when some are active', () => {
        mockMatchMedia(false);
        const onReset = jest.fn();
        render(
            <DocumentsFilterBar
                filters={{ ...emptyDocumentsFilters, favoritesOnly: true }}
                onReset={onReset}
            />
        );

        fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_FILTERS_CLEAR' }));

        expect(onReset).toHaveBeenCalledTimes(1);
    });

    it('collapses into a drawer with an active count below the phone breakpoint', () => {
        mockMatchMedia(true);
        const onChange = jest.fn();
        render(
            <DocumentsFilterBar
                filters={{ ...emptyDocumentsFilters, favoritesOnly: true }}
                onChange={onChange}
            />
        );

        expect(screen.queryByRole('button', { name: 'DOCUMENTS_FILTER_FAVORITES' })).toBeNull();
        expect(screen.getByText('1')).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_FILTERS_BUTTON' }));
        expect(screen.getByText('DOCUMENTS_FILTERS_TITLE')).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_FILTER_FAVORITES' }));
        expect(onChange).toHaveBeenCalledWith({ ...emptyDocumentsFilters, favoritesOnly: false });

        fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_FILTERS_CLOSE' }));
    });
});
