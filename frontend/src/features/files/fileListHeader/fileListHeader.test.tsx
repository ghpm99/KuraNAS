import { fireEvent, render, screen } from '@testing-library/react';
import FileListHeader from './fileListHeader';

describe('FileListHeader', () => {
    it('renders the column headers without any provider or props', () => {
        render(<FileListHeader />);

        expect(screen.getByRole('columnheader', { name: 'NAME' })).toHaveAttribute(
            'aria-sort',
            'ascending'
        );
        expect(screen.getByRole('columnheader', { name: 'SIZE' })).toHaveAttribute('aria-sort', 'none');
        expect(screen.getByRole('columnheader', { name: 'MODIFIED' })).toHaveAttribute(
            'aria-sort',
            'none'
        );
        expect(screen.getByRole('columnheader', { name: 'TYPE' })).not.toHaveAttribute('aria-sort');
        fireEvent.click(screen.getByRole('button', { name: 'SIZE' }));
    });

    it('sorts a new column ascending', () => {
        const onSortChange = jest.fn();
        render(<FileListHeader sort={{ key: 'name', order: 'desc' }} onSortChange={onSortChange} />);

        fireEvent.click(screen.getByRole('button', { name: 'SIZE' }));
        expect(onSortChange).toHaveBeenLastCalledWith({ key: 'size', order: 'asc' });

        fireEvent.click(screen.getByRole('button', { name: 'MODIFIED' }));
        expect(onSortChange).toHaveBeenLastCalledWith({ key: 'updated_at', order: 'asc' });
    });

    it('toggles the order when the active column is clicked again', () => {
        const onSortChange = jest.fn();
        const { rerender } = render(
            <FileListHeader sort={{ key: 'size', order: 'asc' }} onSortChange={onSortChange} />
        );

        expect(screen.getByRole('columnheader', { name: 'SIZE' })).toHaveAttribute(
            'aria-sort',
            'ascending'
        );
        fireEvent.click(screen.getByRole('button', { name: 'SIZE' }));
        expect(onSortChange).toHaveBeenLastCalledWith({ key: 'size', order: 'desc' });

        rerender(<FileListHeader sort={{ key: 'size', order: 'desc' }} onSortChange={onSortChange} />);
        expect(screen.getByRole('columnheader', { name: 'SIZE' })).toHaveAttribute(
            'aria-sort',
            'descending'
        );
        expect(screen.getByTestId('sort-arrow')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: 'SIZE' }));
        expect(onSortChange).toHaveBeenLastCalledWith({ key: 'size', order: 'asc' });
    });

    it('does not make the type column clickable', () => {
        render(<FileListHeader />);

        expect(screen.queryByRole('button', { name: 'TYPE' })).toBeNull();
    });
});
