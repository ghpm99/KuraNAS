import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import PlaylistTrackMoveButtons from './PlaylistTrackMoveButtons';

describe('PlaylistTrackMoveButtons', () => {
    it('renders without any backend', () => {
        expect(() =>
            renderWithoutBackend(
                <PlaylistTrackMoveButtons
                    canMoveUp
                    canMoveDown
                    onMoveUp={jest.fn()}
                    onMoveDown={jest.fn()}
                />
            )
        ).not.toThrow();
    });

    it('moves in the requested direction without bubbling the click', () => {
        const onMoveUp = jest.fn();
        const onMoveDown = jest.fn();
        const onParentClick = jest.fn();
        renderWithoutBackend(
            <div onClick={onParentClick}>
                <PlaylistTrackMoveButtons
                    canMoveUp
                    canMoveDown
                    onMoveUp={onMoveUp}
                    onMoveDown={onMoveDown}
                />
            </div>
        );

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_PLAYLIST_MOVE_UP' }));
        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_PLAYLIST_MOVE_DOWN' }));

        expect(onMoveUp).toHaveBeenCalledTimes(1);
        expect(onMoveDown).toHaveBeenCalledTimes(1);
        expect(onParentClick).not.toHaveBeenCalled();
    });

    it('disables the directions that are not available', () => {
        renderWithoutBackend(
            <PlaylistTrackMoveButtons
                canMoveUp={false}
                canMoveDown={false}
                onMoveUp={jest.fn()}
                onMoveDown={jest.fn()}
            />
        );

        expect(screen.getByRole('button', { name: 'MUSIC_PLAYLIST_MOVE_UP' })).toBeDisabled();
        expect(screen.getByRole('button', { name: 'MUSIC_PLAYLIST_MOVE_DOWN' })).toBeDisabled();
    });
});
