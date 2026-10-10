import { screen, waitFor } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import VideoFolderBrowser from './VideoFolderBrowser';

describe('VideoFolderBrowser without a backend', () => {
    it('mounts with the roots crumb and survives failing requests', async () => {
        renderWithoutBackend(<VideoFolderBrowser onPlayVideo={jest.fn()} />);

        expect(screen.getAllByRole('button')).toHaveLength(1);
        await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
    });
});
