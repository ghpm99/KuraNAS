import { fireEvent, render, screen } from '@testing-library/react';
import FileListRow from './fileListRow';

describe('FileListRow', () => {
    it('renders title and metadata without optional handlers', () => {
        render(<FileListRow title="Doc" href="/files/Doc" sizeText="1 KB" modifiedText="Mar 10" typeText="TXT" thumbnail="/t.png" onClick={jest.fn()} />);

        expect(screen.getByText('TXT')).toBeInTheDocument();
        expect(screen.getByText('1 KB')).toBeInTheDocument();
        expect(screen.getByText('Mar 10')).toBeInTheDocument();
        expect(screen.getByRole('link', { name: 'Doc' })).toHaveAttribute('href', '/files/Doc');
        expect(screen.getByText('1 KB · Mar 10')).toBeInTheDocument();
        expect(screen.queryByRole('checkbox')).toBeNull();
        expect(screen.getByText('☆')).toBeInTheDocument();
    });

    it('shows the secondary text only when provided', () => {
        const { rerender } = render(
            <FileListRow title="Doc" href="/files/Doc" sizeText="1 KB" modifiedText="Mar 10" typeText="TXT" thumbnail="/t.png" onClick={jest.fn()} />
        );
        expect(screen.queryByText('/library/docs')).toBeNull();

        rerender(
            <FileListRow
                title="Doc"
                href="/files/Doc" sizeText="1 KB" modifiedText="Mar 10" typeText="TXT"
                secondaryText="/library/docs"
                thumbnail="/t.png"
                onClick={jest.fn()}
            />
        );
        expect(screen.getByText('/library/docs')).toBeInTheDocument();
    });

    it('shows a filled star and the cold indicator when applicable', () => {
        render(
            <FileListRow
                title="Doc"
                href="/files/Doc" sizeText="1 KB" modifiedText="Mar 10" typeText="TXT"
                thumbnail="/t.png"
                onClick={jest.fn()}
                starred
                isCold
            />
        );

        expect(screen.getByText('★')).toBeInTheDocument();
        expect(screen.getByRole('img', { name: 'FILE_TIER_COLD_INDICATOR' })).toBeInTheDocument();
    });

    it('wires open, star, checkbox, menu button and context menu', () => {
        const onClick = jest.fn();
        const onClickStar = jest.fn();
        const onToggleSelection = jest.fn();
        const onOpenMenu = jest.fn();
        const onContextMenu = jest.fn();
        render(
            <FileListRow
                title="Doc"
                href="/files/Doc" sizeText="1 KB" modifiedText="Mar 10" typeText="TXT"
                thumbnail="/t.png"
                onClick={onClick}
                onClickStar={onClickStar}
                onToggleSelection={onToggleSelection}
                onOpenMenu={onOpenMenu}
                onContextMenu={onContextMenu}
                isSelected
            />
        );

        fireEvent.click(screen.getByRole('link', { name: 'Doc' }));
        fireEvent.click(screen.getByText('☆'));
        fireEvent.click(screen.getByRole('checkbox', { name: 'FILES_SELECT_ITEM' }));
        fireEvent.click(screen.getByRole('button', { name: 'FILES_ITEM_MENU' }));
        fireEvent.contextMenu(screen.getByRole('link', { name: 'Doc' }));

        expect(onClick).toHaveBeenCalledTimes(1);
        expect(onClickStar).toHaveBeenCalledTimes(1);
        expect(onToggleSelection).toHaveBeenCalledTimes(1);
        expect(onOpenMenu).toHaveBeenCalledTimes(1);
        expect(onContextMenu).toHaveBeenCalledTimes(1);
    });

    it('tags the link and reports focus for the roving tab stop', () => {
        const onFocusItem = jest.fn();
        render(
            <FileListRow
                title="Doc"
                href="/files/Doc"
                fileId={4}
                isTabStop={false}
                onFocusItem={onFocusItem}
                sizeText="1 KB"
                modifiedText="Mar 10"
                typeText="TXT"
                thumbnail="/t.png"
                onClick={jest.fn()}
            />
        );

        const link = screen.getByRole('link', { name: 'Doc' });
        expect(link).toHaveAttribute('data-file-id', '4');
        expect(link).toHaveAttribute('tabindex', '-1');
        fireEvent.focus(link);
        expect(onFocusItem).toHaveBeenCalledTimes(1);
    });
});
