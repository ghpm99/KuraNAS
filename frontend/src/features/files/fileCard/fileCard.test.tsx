import { fireEvent, render, screen } from '@testing-library/react';
import FileCard from './fileCard';

describe('components/fileCard', () => {
    it('renders card with thumbnail, metadata and handlers', () => {
        const onClick = jest.fn();
        const onClickStar = jest.fn();
        render(
            <FileCard
                title="Photo"
                metadata="jpg - 1 MB"
                thumbnail="/photo.jpg"
                onClick={onClick}
                starred
                onClickStar={onClickStar}
            />
        );

        expect(screen.getByText('Photo')).toBeInTheDocument();
        expect(screen.getByText('jpg - 1 MB')).toBeInTheDocument();
        fireEvent.click(screen.getByText('Photo'));
        expect(onClick).toHaveBeenCalled();
        fireEvent.click(screen.getAllByRole('button')[1]!);
        expect(onClickStar).toHaveBeenCalled();
    });

    it('uses placeholder thumbnail fallback when image is empty', () => {
        render(<FileCard title="No Image" metadata="meta" thumbnail="" onClick={jest.fn()} />);
        expect(screen.getByAltText('No Image')).toHaveAttribute('src', '/placeholder.svg');
    });

    it('shows the cold indicator only for cold files', () => {
        const { rerender } = render(
            <FileCard title="Doc" metadata="meta" thumbnail="" onClick={jest.fn()} />
        );
        expect(screen.queryByRole('img', { name: 'FILE_TIER_COLD_INDICATOR' })).toBeNull();

        rerender(<FileCard title="Doc" metadata="meta" thumbnail="" onClick={jest.fn()} isCold />);
        expect(screen.getByRole('img', { name: 'FILE_TIER_COLD_INDICATOR' })).toBeInTheDocument();
    });
});

describe('fileCard selection and menu', () => {
    it('renders without selection or menu handlers', () => {
        render(<FileCard title="Plain" metadata="meta" thumbnail="" onClick={jest.fn()} />);
        expect(screen.queryByRole('checkbox')).toBeNull();
        expect(screen.queryByLabelText('FILES_ITEM_MENU')).toBeNull();
    });

    it('toggles selection from the checkbox without opening the card', () => {
        const onClick = jest.fn();
        const onToggleSelection = jest.fn();
        render(
            <FileCard
                title="Doc"
                metadata="meta"
                thumbnail=""
                onClick={onClick}
                onToggleSelection={onToggleSelection}
                isSelected
                isCold
            />
        );

        const checkbox = screen.getByRole('checkbox', { name: 'FILES_SELECT_ITEM' });
        expect(checkbox).toBeChecked();
        fireEvent.click(checkbox);

        expect(onToggleSelection).toHaveBeenCalledTimes(1);
        expect(onClick).not.toHaveBeenCalled();
    });

    it('opens the menu from the button and from the context menu event', () => {
        const onOpenMenu = jest.fn();
        const onContextMenu = jest.fn();
        render(
            <FileCard
                title="Doc"
                metadata="meta"
                thumbnail=""
                onClick={jest.fn()}
                onOpenMenu={onOpenMenu}
                onContextMenu={onContextMenu}
                isSelectionActive
            />
        );

        fireEvent.click(screen.getByRole('button', { name: 'FILES_ITEM_MENU' }));
        expect(onOpenMenu).toHaveBeenCalledTimes(1);

        fireEvent.contextMenu(screen.getByText('Doc'));
        expect(onContextMenu).toHaveBeenCalledTimes(1);
    });
});
