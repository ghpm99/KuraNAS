import { screen, waitFor } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import VideoMovieBrowser from './VideoMovieBrowser';

describe('VideoMovieBrowser without a backend', () => {
    it('mounts with the sort select and survives failing requests', async () => {
        renderWithoutBackend(<VideoMovieBrowser onPlayVideo={jest.fn()} />);

        expect(screen.getByRole('combobox')).toBeInTheDocument();
        await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
    });
});
