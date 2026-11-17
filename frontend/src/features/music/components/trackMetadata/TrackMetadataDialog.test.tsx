import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AxiosError, AxiosHeaders } from 'axios';
import type { IMusicData } from '@/types/music';
import { updateTrackMetadata } from '@/service/music';
import TrackMetadataDialog from './TrackMetadataDialog';

jest.mock('@/service/music', () => ({
    ...jest.requireActual('@/service/music'),
    getMusicArtists: jest.fn().mockResolvedValue({ items: [] }),
    getMusicGenres: jest.fn().mockResolvedValue({ items: [] }),
    updateTrackMetadata: jest.fn(),
}));

const mockedUpdate = updateTrackMetadata as jest.Mock;

const track = {
    id: 7,
    name: 'seven.mp3',
    metadata: {
        title: 'Seven',
        artist: 'Band',
        album: 'Hits',
        album_artist: 'Band',
        genre: 'Rock; Pop',
        year: '1999',
        disc_number: '1/2',
        track_number: '3',
    },
} as IMusicData;

const buildBadRequest = (data: unknown) =>
    new AxiosError('bad', undefined, undefined, undefined, {
        status: 400,
        statusText: '',
        data,
        headers: {},
        config: { headers: new AxiosHeaders() },
    });

const renderDialog = (onClose = jest.fn()) => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
        <QueryClientProvider client={queryClient}>
            <TrackMetadataDialog track={track} onClose={onClose} />
        </QueryClientProvider>
    );
    return onClose;
};

const typeInto = (label: string, text: string) =>
    fireEvent.change(screen.getByLabelText(label), { target: { value: text } });

const clickSave = () =>
    fireEvent.click(screen.getByRole('button', { name: 'MUSIC_METADATA_SAVE' }));

describe('TrackMetadataDialog', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedUpdate.mockResolvedValue(track);
    });

    it('starts from the current values with save disabled', () => {
        renderDialog();

        expect(screen.getByLabelText('MUSIC_METADATA_FIELD_TITLE')).toHaveValue('Seven');
        expect(screen.getByLabelText('MUSIC_METADATA_FIELD_DISC_NUMBER')).toHaveValue('1');
        expect(screen.getByText('Rock')).toBeInTheDocument();
        expect(screen.getByText('Pop')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'MUSIC_METADATA_SAVE' })).toBeDisabled();
    });

    it('sends only the changed fields and closes on success', async () => {
        const onClose = renderDialog();

        typeInto('MUSIC_METADATA_FIELD_TITLE', 'Seven Nation');
        typeInto('MUSIC_METADATA_FIELD_YEAR', '2001');
        clickSave();

        await waitFor(() => expect(onClose).toHaveBeenCalled());
        expect(mockedUpdate).toHaveBeenCalledWith(7, { title: 'Seven Nation', year: 2001 });
    });

    it('sends null for a cleared field', async () => {
        renderDialog();

        typeInto('MUSIC_METADATA_FIELD_ALBUM', '');
        typeInto('MUSIC_METADATA_FIELD_TRACK_NUMBER', '');
        clickSave();

        await waitFor(() =>
            expect(mockedUpdate).toHaveBeenCalledWith(7, { album: null, track_number: null })
        );
    });

    it('sends null when every genre chip is removed', async () => {
        renderDialog();

        const genresInput = screen.getByLabelText('MUSIC_METADATA_FIELD_GENRES');
        fireEvent.keyDown(genresInput, { key: 'Backspace' });
        fireEvent.keyDown(genresInput, { key: 'Backspace' });
        clickSave();

        await waitFor(() => expect(mockedUpdate).toHaveBeenCalledWith(7, { genres: null }));
    });

    it('refuses a non numeric year on the client without calling the server', () => {
        renderDialog();

        typeInto('MUSIC_METADATA_FIELD_YEAR', 'abcd');
        clickSave();

        expect(screen.getByText('MUSIC_METADATA_NUMBER_INVALID')).toBeInTheDocument();
        expect(mockedUpdate).not.toHaveBeenCalled();
    });

    it('shows the server message on the refused field and keeps the dialog open', async () => {
        mockedUpdate.mockRejectedValue(
            buildBadRequest({ error: 'Ano fora do intervalo', field: 'year' })
        );
        const onClose = renderDialog();

        typeInto('MUSIC_METADATA_FIELD_YEAR', '12');
        clickSave();

        expect(await screen.findByText('Ano fora do intervalo')).toBeInTheDocument();
        expect(screen.getByLabelText('MUSIC_METADATA_FIELD_YEAR')).toHaveValue('12');
        expect(screen.getByLabelText('MUSIC_METADATA_FIELD_YEAR')).toHaveAttribute(
            'aria-invalid',
            'true'
        );
        expect(onClose).not.toHaveBeenCalled();
    });

    it('shows an error without field at the top of the dialog', async () => {
        mockedUpdate.mockRejectedValue(new Error('offline'));
        const onClose = renderDialog();

        typeInto('MUSIC_METADATA_FIELD_TITLE', 'Other');
        clickSave();

        expect(await screen.findByRole('alert')).toHaveTextContent('MUSIC_METADATA_SAVE_FAILED');
        expect(screen.getByLabelText('MUSIC_METADATA_FIELD_TITLE')).toHaveValue('Other');
        expect(onClose).not.toHaveBeenCalled();
    });

    it('shows the server message at the top when the 400 names no field', async () => {
        mockedUpdate.mockRejectedValue(buildBadRequest({ error: 'Requisição inválida' }));
        renderDialog();

        typeInto('MUSIC_METADATA_FIELD_TITLE', 'Other');
        clickSave();

        expect(await screen.findByRole('alert')).toHaveTextContent('Requisição inválida');
    });
});
