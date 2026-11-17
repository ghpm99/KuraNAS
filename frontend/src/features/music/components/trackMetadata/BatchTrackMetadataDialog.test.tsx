import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AxiosError, AxiosHeaders } from 'axios';
import { updateTracksMetadata } from '@/service/music';
import type { BatchTrackMetadataSelector } from '@/types/music';
import BatchTrackMetadataDialog from './BatchTrackMetadataDialog';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, options?: Record<string, string>) =>
            [key, ...Object.values(options ?? {})].join(' '),
    }),
}));

jest.mock('@/service/music', () => ({
    ...jest.requireActual('@/service/music'),
    getMusicArtists: jest.fn().mockResolvedValue({ items: [] }),
    getMusicGenres: jest.fn().mockResolvedValue({ items: [] }),
    updateTracksMetadata: jest.fn(),
}));

const mockedUpdate = updateTracksMetadata as jest.Mock;

const albumSelector: BatchTrackMetadataSelector = { albumKey: 'band::hits' };

const buildBadRequest = (data: unknown) =>
    new AxiosError('bad', undefined, undefined, undefined, {
        status: 400,
        statusText: '',
        data,
        headers: {},
        config: { headers: new AxiosHeaders() },
    });

const renderDialog = (selector: BatchTrackMetadataSelector = albumSelector) => {
    const onClose = jest.fn();
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
        <QueryClientProvider client={queryClient}>
            <BatchTrackMetadataDialog selector={selector} onClose={onClose} />
        </QueryClientProvider>
    );
    return onClose;
};

const typeInto = (label: string, text: string) =>
    fireEvent.change(screen.getByLabelText(label), { target: { value: text } });

const clickButton = (name: string) => fireEvent.click(screen.getByRole('button', { name }));

describe('BatchTrackMetadataDialog', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedUpdate.mockResolvedValue({ matched: 3, updated: 3, failed: [] });
    });

    it('titles the dialog after the album or the folder', () => {
        renderDialog({ folderKey: '/music/rock' });

        expect(screen.getByText('MUSIC_EDIT_FOLDER')).toBeInTheDocument();
    });

    it('sends only the filled fields in the simulation and shows the track count', async () => {
        renderDialog();

        typeInto('MUSIC_METADATA_FIELD_ALBUM_ARTIST', '  Band  ');
        typeInto('MUSIC_METADATA_FIELD_YEAR', '1999');
        typeInto('MUSIC_METADATA_FIELD_TITLE', '   ');
        clickButton('MUSIC_BATCH_METADATA_REVIEW');

        expect(await screen.findByText('MUSIC_BATCH_METADATA_WILL_CHANGE 3')).toBeInTheDocument();
        expect(mockedUpdate).toHaveBeenCalledTimes(1);
        expect(mockedUpdate).toHaveBeenCalledWith(
            albumSelector,
            { album_artist: 'Band', year: 1999 },
            { isDryRun: true }
        );
    });

    it('keeps review disabled while nothing is filled', () => {
        renderDialog();

        typeInto('MUSIC_METADATA_FIELD_TITLE', '   ');

        expect(screen.getByRole('button', { name: 'MUSIC_BATCH_METADATA_REVIEW' })).toBeDisabled();
    });

    it('refuses a non numeric year on the client without calling the server', () => {
        renderDialog();

        typeInto('MUSIC_METADATA_FIELD_YEAR', 'abcd');
        clickButton('MUSIC_BATCH_METADATA_REVIEW');

        expect(screen.getByText('MUSIC_METADATA_NUMBER_INVALID')).toBeInTheDocument();
        expect(mockedUpdate).not.toHaveBeenCalled();
    });

    it('blocks confirmation when the simulation matches no track', async () => {
        mockedUpdate.mockResolvedValue({ matched: 0, updated: 0, failed: [] });
        renderDialog();

        typeInto('MUSIC_METADATA_FIELD_TITLE', 'Other');
        clickButton('MUSIC_BATCH_METADATA_REVIEW');

        expect(await screen.findByText('MUSIC_BATCH_METADATA_NO_MATCH')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'MUSIC_BATCH_METADATA_CONFIRM' })).toBeDisabled();
    });

    it('goes back to the fields keeping what was typed', async () => {
        renderDialog();

        typeInto('MUSIC_METADATA_FIELD_TITLE', 'Other');
        clickButton('MUSIC_BATCH_METADATA_REVIEW');
        await screen.findByText('MUSIC_BATCH_METADATA_WILL_CHANGE 3');
        clickButton('MUSIC_BATCH_METADATA_BACK');

        expect(screen.getByLabelText('MUSIC_METADATA_FIELD_TITLE')).toHaveValue('Other');
    });

    it('confirms without simulation and closes when every track was updated', async () => {
        const onClose = renderDialog();

        typeInto('MUSIC_METADATA_FIELD_TITLE', 'Other');
        clickButton('MUSIC_BATCH_METADATA_REVIEW');
        await screen.findByText('MUSIC_BATCH_METADATA_WILL_CHANGE 3');
        clickButton('MUSIC_BATCH_METADATA_CONFIRM');

        await waitFor(() => expect(onClose).toHaveBeenCalled());
        expect(mockedUpdate).toHaveBeenLastCalledWith(
            albumSelector,
            { title: 'Other' },
            { isDryRun: false }
        );
    });

    it('lists the failed tracks with the server error and stays open', async () => {
        mockedUpdate
            .mockResolvedValueOnce({ matched: 3, updated: 0, failed: [] })
            .mockResolvedValueOnce({
                matched: 3,
                updated: 2,
                failed: [{ fileId: 41, error: 'Arquivo somente leitura' }],
            });
        const onClose = renderDialog();

        typeInto('MUSIC_METADATA_FIELD_TITLE', 'Other');
        clickButton('MUSIC_BATCH_METADATA_REVIEW');
        await screen.findByText('MUSIC_BATCH_METADATA_WILL_CHANGE 3');
        clickButton('MUSIC_BATCH_METADATA_CONFIRM');

        expect(await screen.findByText('MUSIC_BATCH_METADATA_UPDATED 2')).toBeInTheDocument();
        expect(screen.getByText('MUSIC_BATCH_METADATA_FAILED_FILE 41')).toBeInTheDocument();
        expect(screen.getByText('Arquivo somente leitura')).toBeInTheDocument();
        expect(onClose).not.toHaveBeenCalled();
        clickButton('CLOSE');
        expect(onClose).toHaveBeenCalled();
    });

    it('shows the server message on the refused field and returns to the fields', async () => {
        mockedUpdate.mockRejectedValue(
            buildBadRequest({ error: 'Ano fora do intervalo', field: 'year' })
        );
        renderDialog();

        typeInto('MUSIC_METADATA_FIELD_YEAR', '12');
        clickButton('MUSIC_BATCH_METADATA_REVIEW');

        expect(await screen.findByText('Ano fora do intervalo')).toBeInTheDocument();
        expect(screen.getByLabelText('MUSIC_METADATA_FIELD_YEAR')).toHaveAttribute(
            'aria-invalid',
            'true'
        );
    });

    it('shows an error without field at the top and keeps the current step', async () => {
        mockedUpdate.mockRejectedValue(new Error('offline'));
        renderDialog();

        typeInto('MUSIC_METADATA_FIELD_TITLE', 'Other');
        clickButton('MUSIC_BATCH_METADATA_REVIEW');

        expect(await screen.findByRole('alert')).toHaveTextContent('MUSIC_BATCH_METADATA_FAILED');
        expect(screen.getByLabelText('MUSIC_METADATA_FIELD_TITLE')).toHaveValue('Other');
    });
});
