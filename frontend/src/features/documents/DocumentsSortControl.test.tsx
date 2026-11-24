import { fireEvent, render, screen, within } from '@testing-library/react';
import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import DocumentsSortControl from './DocumentsSortControl';

describe('DocumentsSortControl', () => {
    it('renders without props and without a backend', () => {
        expectRendersWithoutBackend(<DocumentsSortControl />);
    });

    it('toggles the order keeping the field', () => {
        const onChange = jest.fn();
        render(
            <DocumentsSortControl sortState={{ sort: 'size', order: 'desc' }} onChange={onChange} />
        );

        fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_SORT_ORDER_DESCENDING' }));

        expect(onChange).toHaveBeenCalledWith({ sort: 'size', order: 'asc' });
    });

    it('switches the field to its natural order', () => {
        const onChange = jest.fn();
        render(<DocumentsSortControl onChange={onChange} />);

        fireEvent.mouseDown(screen.getByRole('combobox'));
        fireEvent.click(within(screen.getByRole('listbox')).getByText('DOCUMENTS_SORT_NAME'));

        expect(onChange).toHaveBeenCalledWith({ sort: 'name', order: 'asc' });
    });
});

describe('DocumentsSortControl while searching', () => {
    it('shows Relevância locked and hides the order toggle', () => {
        render(<DocumentsSortControl isRelevanceLocked />);

        expect(screen.getByRole('combobox')).toHaveAttribute('aria-disabled', 'true');
        expect(screen.getByText('FILES_SEARCH_SORT_RELEVANCE')).toBeInTheDocument();
        expect(screen.queryByRole('button')).toBeNull();
    });
});
