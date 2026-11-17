import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AxiosError, AxiosHeaders } from 'axios';
import { renameMusicGenre } from '@/service/music';
import RenameGenreDialog from './RenameGenreDialog';

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
    getMusicGenres: jest.fn().mockResolvedValue({
        items: [
            { key: 'pop', genre: 'Pop', track_count: 40 },
            { key: 'rock', genre: 'Rock', track_count: 300 },
        ],
    }),
    renameMusicGenre: jest.fn(),
}));

const mockedRename = renameMusicGenre as jest.Mock;

const buildAxiosFailure = (status: number, data: unknown) =>
    new AxiosError('failed', undefined, undefined, undefined, {
        status,
        statusText: '',
        data,
        headers: {},
        config: { headers: new AxiosHeaders() },
    });

const renderDialog = () => {
    const onRenamed = jest.fn();
    const onClose = jest.fn();
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <RenameGenreDialog
                genreKey="pop"
                genreName="Pop"
                trackCount={40}
                onRenamed={onRenamed}
                onClose={onClose}
            />
        </QueryClientProvider>
    );
    return { onRenamed, onClose };
};

const nameField = () => screen.getByLabelText('MUSIC_GENRE_RENAME_FIELD_NAME');

const typeName = (typedName: string) =>
    fireEvent.change(nameField(), { target: { value: typedName } });

const clickButton = (name: string) => fireEvent.click(screen.getByRole('button', { name }));

const pickSuggestion = async (suggestionName: string) => {
    fireEvent.keyDown(nameField(), { key: 'ArrowDown' });
    fireEvent.click(await screen.findByRole('option', { name: suggestionName }));
};

describe('RenameGenreDialog', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedRename.mockResolvedValue({
            key: 'indie-pop',
            name: 'Indie Pop',
            trackCount: 40,
            merged: false,
        });
    });

    it('states how many tracks move before sending a plain rename', async () => {
        const { onRenamed, onClose } = renderDialog();

        typeName('Indie Pop');
        clickButton('MUSIC_GENRE_RENAME_REVIEW');

        expect(
            screen.getByText('MUSIC_GENRE_RENAME_WILL_RENAME 40 Pop Indie Pop 40')
        ).toBeVisible();
        expect(mockedRename).not.toHaveBeenCalled();

        clickButton('MUSIC_GENRE_RENAME_CONFIRM');

        await waitFor(() => expect(onRenamed).toHaveBeenCalledWith('indie-pop'));
        expect(mockedRename).toHaveBeenCalledTimes(1);
        expect(mockedRename).toHaveBeenCalledWith('pop', 'Indie Pop');
        expect(onClose).toHaveBeenCalledTimes(1);
    });

    it('announces a merge with the summed total when an existing genre is picked', async () => {
        renderDialog();

        await pickSuggestion('Rock');
        clickButton('MUSIC_GENRE_RENAME_REVIEW');

        expect(screen.getByText('MUSIC_GENRE_RENAME_WILL_MERGE 40 Pop Rock 340')).toBeVisible();
    });

    it('does not offer the genre being renamed as a merge target', async () => {
        renderDialog();

        fireEvent.keyDown(nameField(), { key: 'ArrowDown' });
        await screen.findByRole('option', { name: 'Rock' });

        expect(screen.queryByRole('option', { name: 'Pop' })).not.toBeInTheDocument();
    });

    it('treats typed text as a rename even when it spells an existing genre, leaving the merge to the server', async () => {
        mockedRename.mockResolvedValue({
            key: 'rock',
            name: 'Rock',
            trackCount: 340,
            merged: true,
        });
        const { onRenamed } = renderDialog();

        await pickSuggestion('Rock');
        typeName('Rock ');
        clickButton('MUSIC_GENRE_RENAME_REVIEW');

        expect(screen.getByText('MUSIC_GENRE_RENAME_WILL_RENAME 40 Pop Rock 40')).toBeVisible();
        clickButton('MUSIC_GENRE_RENAME_CONFIRM');
        await waitFor(() => expect(onRenamed).toHaveBeenCalledWith('rock'));
    });

    it('blocks review while the name is blank', () => {
        renderDialog();

        typeName('   ');

        expect(screen.getByRole('button', { name: 'MUSIC_GENRE_RENAME_REVIEW' })).toBeDisabled();
    });

    it('locks the field and the buttons while the rename is in flight', async () => {
        mockedRename.mockReturnValue(new Promise(() => undefined));
        renderDialog();

        clickButton('MUSIC_GENRE_RENAME_REVIEW');
        clickButton('MUSIC_GENRE_RENAME_CONFIRM');

        await waitFor(() =>
            expect(
                screen.getByRole('button', { name: 'MUSIC_GENRE_RENAME_CONFIRM' })
            ).toBeDisabled()
        );
        expect(screen.getByRole('button', { name: 'MUSIC_GENRE_RENAME_BACK' })).toBeDisabled();
        expect(nameField()).toBeDisabled();
    });

    it('shows the server text on the name field when the name is refused', async () => {
        mockedRename.mockRejectedValue(
            buildAxiosFailure(400, { error: 'Nome inválido', field: 'name' })
        );
        const { onClose } = renderDialog();

        typeName('A;B');
        clickButton('MUSIC_GENRE_RENAME_REVIEW');
        clickButton('MUSIC_GENRE_RENAME_CONFIRM');

        expect(await screen.findByText('Nome inválido')).toBeVisible();
        expect(nameField()).toHaveAttribute('aria-invalid', 'true');
        expect(nameField()).toHaveValue('A;B');
        expect(onClose).not.toHaveBeenCalled();
    });

    it('shows other server failures as an alert and keeps the dialog open', async () => {
        mockedRename.mockRejectedValue(buildAxiosFailure(404, { error: 'Gênero não encontrado' }));
        const { onRenamed, onClose } = renderDialog();

        clickButton('MUSIC_GENRE_RENAME_REVIEW');
        clickButton('MUSIC_GENRE_RENAME_CONFIRM');

        expect(await screen.findByRole('alert')).toHaveTextContent('Gênero não encontrado');
        expect(onRenamed).not.toHaveBeenCalled();
        expect(onClose).not.toHaveBeenCalled();
    });

    it('falls back to the translated failure text when the server sends none', async () => {
        mockedRename.mockRejectedValue(new Error('offline'));
        renderDialog();

        clickButton('MUSIC_GENRE_RENAME_REVIEW');
        clickButton('MUSIC_GENRE_RENAME_CONFIRM');

        expect(await screen.findByRole('alert')).toHaveTextContent('MUSIC_GENRE_RENAME_FAILED');
    });

    it('goes back to the field keeping what was typed', () => {
        renderDialog();

        typeName('Indie Pop');
        clickButton('MUSIC_GENRE_RENAME_REVIEW');
        clickButton('MUSIC_GENRE_RENAME_BACK');

        expect(nameField()).toHaveValue('Indie Pop');
    });
});
