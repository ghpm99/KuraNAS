import { screen, waitFor, within } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import VideoSeriesDetailView from './VideoSeriesDetailView';

describe('VideoSeriesDetailView without mocks', () => {
    it('mounts against an absent backend and offers a retry for the failed header', async () => {
        renderWithoutBackend(
            <VideoSeriesDetailView
                playlist={{ id: 3, name: 'Bare Show', item_count: 0 }}
                onBack={jest.fn()}
                onOpenVideo={jest.fn()}
            />
        );

        await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
        expect(within(screen.getByRole('alert')).getByRole('button')).toBeInTheDocument();
    });
});
