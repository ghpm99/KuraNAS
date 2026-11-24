import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import type { VideoFileDto } from '@/service/videoPlayback';
import VideoMovieGrid from './VideoMovieGrid';

const movie = { id: 3, name: 'Filme.mkv', format: 'mkv', path: '', parent_path: '', size: 1 };

describe('VideoMovieGrid classification correction', () => {
    it('renders without a backend and offers only the classification correction', () => {
        renderWithoutBackend(
            <VideoMovieGrid movies={[movie as VideoFileDto]} onPlayVideo={jest.fn()} />
        );

        fireEvent.click(screen.getByRole('button', { name: 'VIDEO_CORRECTION_MENU' }));

        expect(
            screen.getByRole('menuitem', { name: 'VIDEO_CORRECT_CLASSIFICATION' })
        ).toBeInTheDocument();
        expect(screen.queryByRole('menuitem', { name: 'VIDEO_CORRECT_EPISODE' })).toBeNull();
    });
});
