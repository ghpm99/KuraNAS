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

    it('shows the secondary text under the title only when provided', () => {
        const { rerender } = render(
            <FileCard title="Doc" metadata="meta" thumbnail="" onClick={jest.fn()} />
        );
        expect(screen.queryByText('/library/docs')).toBeNull();

        rerender(
            <FileCard
                title="Doc"
                metadata="meta"
                secondaryText="/library/docs"
                thumbnail=""
                onClick={jest.fn()}
            />
        );
        expect(screen.getByText('/library/docs')).toBeInTheDocument();
    });

    it('uses placeholder thumbnail fallback when image is empty', () => {
        const { container } = render(
            <FileCard title="No Image" metadata="meta" thumbnail="" onClick={jest.fn()} />
        );
        expect(container.querySelector('img')).toHaveAttribute('src', '/placeholder.svg');
    });

    it('keeps the thumbnail decorative so the card name is announced once', () => {
        const { container } = render(
            <FileCard title="Photo" metadata="meta" thumbnail="/photo.jpg" onClick={jest.fn()} />
        );

        expect(container.querySelector('img')).toHaveAttribute('alt', '');
    });

    it('labels the favorite toggle with the item name and its pressed state', () => {
        const { rerender } = render(
            <FileCard title="Photo" metadata="meta" thumbnail="" onClick={jest.fn()} />
        );
        expect(
            screen.getByRole('button', { name: 'FILES_FAVORITE_ITEM', pressed: false })
        ).toBeInTheDocument();

        rerender(<FileCard title="Photo" metadata="meta" thumbnail="" onClick={jest.fn()} starred />);
        expect(
            screen.getByRole('button', { name: 'FILES_UNFAVORITE_ITEM', pressed: true })
        ).toBeInTheDocument();
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

    it('renders the action area as a link when an href is given', () => {
        render(
            <FileCard title="Photo" metadata="meta" thumbnail="" href="/files/Photo" onClick={jest.fn()} />
        );

        expect(screen.getByRole('link', { name: /Photo/ })).toHaveAttribute('href', '/files/Photo');
    });

    it('tags the primary link and reports focus for the roving tab stop', () => {
        const onFocusItem = jest.fn();
        render(
            <FileCard
                title="Doc"
                metadata="meta"
                thumbnail=""
                href="/files/doc"
                fileId={9}
                isTabStop={false}
                onFocusItem={onFocusItem}
                onClick={jest.fn()}
            />
        );

        const link = screen.getByRole('link');
        expect(link).toHaveAttribute('data-file-id', '9');
        expect(link).toHaveAttribute('tabindex', '-1');
        fireEvent.focus(link);
        expect(onFocusItem).toHaveBeenCalledTimes(1);
    });
});
