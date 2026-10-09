import { fireEvent, render, screen, within } from '@testing-library/react';
import FilesSortControl from './filesSortControl';

describe('FilesSortControl', () => {
    it('renders without props, providers or service mocks', () => {
        render(<FilesSortControl />);

        expect(screen.getByRole('combobox')).toHaveTextContent('FILES_SORT_NAME');
        expect(
            screen.getByRole('button', { name: 'FILES_SORT_ORDER_ASCENDING' })
        ).toBeInTheDocument();
    });

    it('reports the chosen key keeping the current order', () => {
        const onChange = jest.fn();
        render(<FilesSortControl sort={{ key: 'name', order: 'desc' }} onChange={onChange} />);

        fireEvent.mouseDown(screen.getByRole('combobox'));
        fireEvent.click(within(screen.getByRole('listbox')).getByText('FILES_SORT_SIZE'));

        expect(onChange).toHaveBeenCalledWith({ key: 'size', order: 'desc' });
    });

    it('toggles the order keeping the current key', () => {
        const onChange = jest.fn();
        render(<FilesSortControl sort={{ key: 'updated_at', order: 'asc' }} onChange={onChange} />);

        fireEvent.click(screen.getByRole('button', { name: 'FILES_SORT_ORDER_ASCENDING' }));

        expect(onChange).toHaveBeenCalledWith({ key: 'updated_at', order: 'desc' });
    });

    it('shows the descending label when sorted descending', () => {
        render(<FilesSortControl sort={{ key: 'created_at', order: 'desc' }} />);

        expect(
            screen.getByRole('button', { name: 'FILES_SORT_ORDER_DESCENDING' })
        ).toBeInTheDocument();
    });
});
