import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import DeleteItemsDialog from './deleteItemsDialog';

const items = [{ id: 1 }, { id: 2 }];

const renderOpenDialog = (props: Partial<Parameters<typeof DeleteItemsDialog>[0]> = {}) =>
    render(
        <MemoryRouter>
            <DeleteItemsDialog items={items} isOpen onClose={jest.fn()} onConfirm={jest.fn()} {...props} />
        </MemoryRouter>
    );

describe('DeleteItemsDialog', () => {
    it('renders closed without a router or any provider', () => {
        render(<DeleteItemsDialog items={[]} isOpen={false} onClose={jest.fn()} onConfirm={jest.fn()} />);

        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    it('says the items go to the trash and links to the trash page', () => {
        renderOpenDialog();

        expect(screen.getByText(/FILES_DELETE_TRASH_NOTICE/)).toBeInTheDocument();
        expect(screen.getByRole('link', { name: 'FILES_DELETE_OPEN_TRASH' })).toHaveAttribute(
            'href',
            '/trash'
        );
    });

    it('confirms a regular delete by default', () => {
        const onConfirm = jest.fn();
        renderOpenDialog({ onConfirm });

        expect(screen.getByRole('checkbox', { name: 'FILES_DELETE_PERMANENTLY' })).not.toBeChecked();
        expect(screen.queryByText('FILES_DELETE_PERMANENT_WARNING')).not.toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: 'DELETE' }));

        expect(onConfirm).toHaveBeenCalledWith(items, false);
    });

    it('shows the red warning and confirms a permanent delete when checked', () => {
        const onConfirm = jest.fn();
        renderOpenDialog({ onConfirm });

        fireEvent.click(screen.getByRole('checkbox', { name: 'FILES_DELETE_PERMANENTLY' }));

        expect(screen.getByRole('alert')).toHaveTextContent('FILES_DELETE_PERMANENT_WARNING');
        fireEvent.click(screen.getAllByRole('button', { name: 'FILES_DELETE_PERMANENTLY' })[0]!);
        expect(onConfirm).toHaveBeenCalledWith(items, true);
    });

    it('uses the many-items message for several files and closes the dialog from the trash link', () => {
        const onClose = jest.fn();
        renderOpenDialog({ onClose });

        expect(screen.getByText('FILES_CONFIRM_DELETE_MANY')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('link', { name: 'FILES_DELETE_OPEN_TRASH' }));
        expect(onClose).toHaveBeenCalledTimes(1);
    });
});
