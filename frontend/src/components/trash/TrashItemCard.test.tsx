import { fireEvent, render, screen } from '@testing-library/react';
import TrashItemCard from './TrashItemCard';
import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import type { TrashItem } from '@/types/trash';

const daysFromNow = (dayCount: number): string =>
    new Date(Date.now() + dayCount * 24 * 60 * 60 * 1000 - 60 * 1000).toISOString();

const deletedReport: TrashItem = {
    id: 7,
    name: 'relatorio.pdf',
    is_directory: false,
    original_path: '/data/docs/relatorio.pdf',
    size: 2048,
    deleted_at: '2026-06-11T10:00:00Z',
    expires_at: daysFromNow(12),
};

const renderCard = (item: TrashItem, isBusy = false) => {
    const onRestore = jest.fn();
    const onDeleteForever = jest.fn();
    render(
        <TrashItemCard
            item={item}
            isBusy={isBusy}
            onRestore={onRestore}
            onDeleteForever={onDeleteForever}
        />
    );
    return { onRestore, onDeleteForever };
};

describe('TrashItemCard', () => {
    it('renders an item that only has an id without backend', () => {
        expectRendersWithoutBackend(
            <TrashItemCard
                item={{ id: 1 } as TrashItem}
                isBusy={false}
                onRestore={jest.fn()}
                onDeleteForever={jest.fn()}
            />
        );
    });

    it('shows the name, the origin folder and the days left', () => {
        renderCard(deletedReport);

        expect(screen.getByText('relatorio.pdf')).toBeInTheDocument();
        expect(screen.getByText('TRASH_ORIGIN_LABEL: /data/docs')).toBeInTheDocument();
        expect(screen.getByText('TRASH_EXPIRES_IN_DAYS')).toBeInTheDocument();
        expect(screen.getByRole('img', { name: 'TRASH_KIND_FILE' })).toBeInTheDocument();
    });

    it('marks folders with the folder icon', () => {
        renderCard({ ...deletedReport, name: 'fotos', is_directory: true });

        expect(screen.getByRole('img', { name: 'TRASH_KIND_FOLDER' })).toBeInTheDocument();
    });

    it('uses the singular label on the last day', () => {
        renderCard({ ...deletedReport, expires_at: daysFromNow(1) });

        expect(screen.getByText('TRASH_EXPIRES_IN_ONE_DAY')).toBeInTheDocument();
    });

    it('tells an overdue item leaves on the next cleanup', () => {
        renderCard({ ...deletedReport, expires_at: daysFromNow(-2) });

        expect(screen.getByText('TRASH_EXPIRES_NEXT_CLEANUP')).toBeInTheDocument();
    });

    it('omits the expiry when the server does not send it', () => {
        renderCard({ ...deletedReport, expires_at: undefined });

        expect(screen.queryByText(/TRASH_EXPIRES/)).not.toBeInTheDocument();
    });

    it('asks to restore and to delete the item by id', () => {
        const { onRestore, onDeleteForever } = renderCard(deletedReport);

        fireEvent.click(screen.getByRole('button', { name: /TRASH_RESTORE_BUTTON/ }));
        fireEvent.click(screen.getByRole('button', { name: /TRASH_DELETE_FOREVER_BUTTON/ }));

        expect(onRestore).toHaveBeenCalledWith(7);
        expect(onDeleteForever).toHaveBeenCalledWith(7);
    });

    it('locks both actions while busy', () => {
        renderCard(deletedReport, true);

        expect(screen.getByRole('button', { name: /TRASH_RESTORE_BUTTON/ })).toBeDisabled();
        expect(screen.getByRole('button', { name: /TRASH_DELETE_FOREVER_BUTTON/ })).toBeDisabled();
    });
});
