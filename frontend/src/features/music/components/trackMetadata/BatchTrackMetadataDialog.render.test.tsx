import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import BatchTrackMetadataDialog from './BatchTrackMetadataDialog';

describe('BatchTrackMetadataDialog with no service mock and no backend', () => {
    it('mounts with every field empty and review disabled', () => {
        renderWithoutBackend(
            <BatchTrackMetadataDialog selector={{ albumKey: 'band::hits' }} onClose={jest.fn()} />
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
        expect(screen.getByLabelText('MUSIC_METADATA_FIELD_TITLE')).toHaveValue('');
        expect(screen.getByRole('button', { name: 'MUSIC_BATCH_METADATA_REVIEW' })).toBeDisabled();
    });
});
