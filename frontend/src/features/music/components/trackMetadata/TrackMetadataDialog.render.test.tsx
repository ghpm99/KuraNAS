import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { IMusicData } from '@/types/music';
import TrackMetadataDialog from './TrackMetadataDialog';

describe('TrackMetadataDialog with no service mock and no backend', () => {
    it('mounts, shows every field and keeps save disabled', () => {
        const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });

        render(
            <QueryClientProvider client={queryClient}>
                <TrackMetadataDialog
                    track={{ id: 3, name: 'bare.mp3' } as IMusicData}
                    onClose={jest.fn()}
                />
            </QueryClientProvider>
        );

        [
            'MUSIC_METADATA_FIELD_TITLE',
            'MUSIC_METADATA_FIELD_ARTIST',
            'MUSIC_METADATA_FIELD_ALBUM',
            'MUSIC_METADATA_FIELD_ALBUM_ARTIST',
            'MUSIC_METADATA_FIELD_GENRES',
            'MUSIC_METADATA_FIELD_YEAR',
            'MUSIC_METADATA_FIELD_DISC_NUMBER',
            'MUSIC_METADATA_FIELD_TRACK_NUMBER',
        ].forEach((label) => expect(screen.getByLabelText(label)).toBeInTheDocument());
        expect(screen.getByRole('button', { name: 'MUSIC_METADATA_SAVE' })).toBeDisabled();
    });
});
