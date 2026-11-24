import { fireEvent, render, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import DocumentMenuButton from './DocumentMenuButton';
import DocumentSelectionCheckbox from './DocumentSelectionCheckbox';
import DocumentStarButton from './DocumentStarButton';

describe('document item controls', () => {
    it('render without a backend', () => {
        expect(() =>
            renderWithoutBackend(
                <>
                    <DocumentSelectionCheckbox documentName="a" isSelected={false} onToggle={jest.fn()} />
                    <DocumentStarButton documentName="a" isStarred={false} onToggle={jest.fn()} />
                    <DocumentMenuButton documentName="a" onOpen={jest.fn()} />
                </>
            )
        ).not.toThrow();
    });

    it('reports the clicks', () => {
        const onToggleSelection = jest.fn();
        const onToggleStar = jest.fn();
        const onOpenMenu = jest.fn();
        render(
            <>
                <DocumentSelectionCheckbox documentName="a" isSelected onToggle={onToggleSelection} />
                <DocumentStarButton documentName="a" isStarred onToggle={onToggleStar} />
                <DocumentMenuButton documentName="a" onOpen={onOpenMenu} />
            </>
        );

        fireEvent.click(screen.getByRole('checkbox'));
        fireEvent.click(screen.getByRole('button', { name: 'FILES_UNFAVORITE_ITEM' }));
        fireEvent.click(screen.getByRole('button', { name: 'FILES_ITEM_MENU' }));

        expect(onToggleSelection).toHaveBeenCalledTimes(1);
        expect(onToggleStar).toHaveBeenCalledTimes(1);
        expect(onOpenMenu).toHaveBeenCalledTimes(1);
        expect(screen.getByRole('button', { name: 'FILES_UNFAVORITE_ITEM' })).toHaveAttribute(
            'aria-pressed',
            'true'
        );
    });
});
