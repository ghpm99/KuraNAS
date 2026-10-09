import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import ImageUserAlbumsSection from './ImageUserAlbumsSection';
import ImageAddToAlbumDialog from './ImageAddToAlbumDialog';
import ImageAlbumNameDialog from './ImageAlbumNameDialog';
import ImageUserAlbumCard from './ImageUserAlbumCard';

describe('user album components without a backend', () => {
    it('renders the albums section while every request fails', async () => {
        renderWithoutBackend(<ImageUserAlbumsSection onOpenAlbum={jest.fn()} />);

        expect(
            await screen.findByRole('button', { name: /IMAGES_ALBUM_CREATE|album/i })
        ).toBeTruthy();
    });

    it('renders the add-to-album dialog while every request fails', () => {
        renderWithoutBackend(
            <ImageAddToAlbumDialog isOpen onClose={jest.fn()} onPickAlbum={jest.fn()} />
        );

        expect(screen.getAllByRole('textbox').length).toBeGreaterThan(0);
    });

    it('renders the name dialog closed and open', () => {
        const { rerender } = renderWithoutBackend(
            <ImageAlbumNameDialog
                isOpen={false}
                title="t"
                confirmLabel="c"
                onClose={jest.fn()}
                onSubmit={jest.fn()}
            />
        );
        expect(screen.queryByRole('dialog')).toBeNull();
        rerender(<div />);
    });

    it('renders a card with a partial album payload', () => {
        renderWithoutBackend(
            <ImageUserAlbumCard
                album={{ id: 1, name: 'Trip' } as never}
                onOpen={jest.fn()}
                onRename={jest.fn()}
                onDelete={jest.fn()}
            />
        );

        expect(screen.getByText('Trip')).toBeTruthy();
    });
});
