import { fireEvent, render, screen, within } from '@testing-library/react';
import TrashListingControls from './TrashListingControls';
import type { TrashSort } from './useTrashScreen';
import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';

const newestFirst: TrashSort = { sortKey: 'deleted_at', order: 'desc' };

const renderControls = (sort: TrashSort = newestFirst) => {
    const onSortChange = jest.fn();
    const onSearchInputChange = jest.fn();
    render(
        <TrashListingControls
            sort={sort}
            onSortChange={onSortChange}
            searchInput=""
            onSearchInputChange={onSearchInputChange}
        />
    );
    return { onSortChange, onSearchInputChange };
};

describe('TrashListingControls', () => {
    it('renders without backend', () => {
        expectRendersWithoutBackend(
            <TrashListingControls
                sort={newestFirst}
                onSortChange={jest.fn()}
                searchInput=""
                onSearchInputChange={jest.fn()}
            />
        );
    });

    it('reports the typed search text', () => {
        const { onSearchInputChange } = renderControls();

        fireEvent.change(screen.getByLabelText('TRASH_SEARCH_LABEL'), {
            target: { value: 'fotos' },
        });

        expect(onSearchInputChange).toHaveBeenCalledWith('fotos');
    });

    it('changes the sort key keeping the order', () => {
        const { onSortChange } = renderControls();

        fireEvent.mouseDown(screen.getByRole('combobox'));
        fireEvent.click(within(screen.getByRole('listbox')).getByText('TRASH_SORT_SIZE'));

        expect(onSortChange).toHaveBeenCalledWith({ sortKey: 'size', order: 'desc' });
    });

    it('flips the order keeping the sort key', () => {
        const { onSortChange } = renderControls({ sortKey: 'name', order: 'asc' });

        fireEvent.click(screen.getByRole('button', { name: 'TRASH_SORT_ORDER_ASCENDING' }));

        expect(onSortChange).toHaveBeenCalledWith({ sortKey: 'name', order: 'desc' });
    });
});
