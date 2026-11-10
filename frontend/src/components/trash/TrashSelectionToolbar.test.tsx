import { fireEvent, render, screen } from '@testing-library/react';
import TrashSelectionToolbar from './TrashSelectionToolbar';
import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';

const renderToolbar = (selectedCount: number, isBusy = false) => {
    const toolbarActions = {
        onSelectAll: jest.fn(),
        onClearSelection: jest.fn(),
        onRestoreSelected: jest.fn(),
        onDeleteSelected: jest.fn(),
    };
    render(
        <TrashSelectionToolbar selectedCount={selectedCount} isBusy={isBusy} {...toolbarActions} />
    );
    return toolbarActions;
};

describe('TrashSelectionToolbar', () => {
    it('renders without backend', () => {
        expectRendersWithoutBackend(
            <TrashSelectionToolbar
                selectedCount={0}
                isBusy={false}
                onSelectAll={jest.fn()}
                onClearSelection={jest.fn()}
                onRestoreSelected={jest.fn()}
                onDeleteSelected={jest.fn()}
            />
        );
    });

    it('only offers selecting the page while nothing is selected', () => {
        const { onSelectAll } = renderToolbar(0);

        fireEvent.click(screen.getByRole('button', { name: 'TRASH_SELECT_ALL' }));

        expect(onSelectAll).toHaveBeenCalled();
        expect(screen.queryByText('TRASH_SELECTION_COUNT')).not.toBeInTheDocument();
        expect(
            screen.queryByRole('button', { name: /TRASH_RESTORE_SELECTED_BUTTON/ })
        ).not.toBeInTheDocument();
    });

    it('offers the batch actions once something is selected', () => {
        const { onClearSelection, onRestoreSelected, onDeleteSelected } = renderToolbar(2);

        expect(screen.getByText('TRASH_SELECTION_COUNT')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: /TRASH_RESTORE_SELECTED_BUTTON/ }));
        fireEvent.click(screen.getByRole('button', { name: /TRASH_DELETE_SELECTED_BUTTON/ }));
        fireEvent.click(screen.getByRole('button', { name: 'TRASH_CLEAR_SELECTION' }));

        expect(onRestoreSelected).toHaveBeenCalled();
        expect(onDeleteSelected).toHaveBeenCalled();
        expect(onClearSelection).toHaveBeenCalled();
    });

    it('locks the batch actions while busy', () => {
        renderToolbar(2, true);

        expect(
            screen.getByRole('button', { name: /TRASH_RESTORE_SELECTED_BUTTON/ })
        ).toBeDisabled();
        expect(screen.getByRole('button', { name: /TRASH_DELETE_SELECTED_BUTTON/ })).toBeDisabled();
    });
});
