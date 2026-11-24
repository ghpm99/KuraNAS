import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import * as musicService from '@/service/music';
import { buildServerPage } from '@/shared/test/buildServerPage';
import { PlaylistRules } from '@/types/playlist';
import PlaylistRuleCriteriaSummary from './PlaylistRuleCriteriaSummary';

jest.mock('@/service/music');

const mockedMusic = jest.mocked(musicService);
const pageOf = <ItemType,>(items: ItemType[]) =>
    buildServerPage<ItemType>(items, { page: 1, pageSize: 500, hasNext: false });

const rules: PlaylistRules = {
    match: 'any',
    genres: ['hip hop'],
    artists: ['queen'],
    albums: ['queen::a night at the opera'],
    folders: [],
    year_from: null,
    year_to: null,
    favorites_only: false,
};

describe('PlaylistRuleCriteriaSummary labels', () => {
    it('resolves genre, artist and album keys to their real labels', async () => {
        mockedMusic.getMusicGenres.mockResolvedValue(
            pageOf([{ key: 'hip hop', genre: 'Hip Hop' }]) as never
        );
        mockedMusic.getMusicArtists.mockResolvedValue(
            pageOf([{ key: 'queen', artist: 'Queen' }]) as never
        );
        mockedMusic.getMusicAlbums.mockResolvedValue(
            pageOf([
                {
                    key: 'queen::a night at the opera',
                    album: 'A Night at the Opera',
                    artist: 'Queen',
                },
            ]) as never
        );

        renderWithoutBackend(<PlaylistRuleCriteriaSummary rules={rules} />);

        expect(await screen.findByText('Hip Hop')).toBeInTheDocument();
        expect(await screen.findByText('Queen')).toBeInTheDocument();
        expect(await screen.findByText('A Night at the Opera - Queen')).toBeInTheDocument();
    });
});
