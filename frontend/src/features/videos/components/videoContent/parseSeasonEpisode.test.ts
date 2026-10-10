import { parseSeasonEpisode } from './parseSeasonEpisode';

describe('parseSeasonEpisode', () => {
    it.each(['Movie 1920x1080', '1280x720', 'step 2', 'Movie 1920x1080 S01E02x'])(
        'ignores %s',
        (name) => {
            expect(parseSeasonEpisode(name)).toBeNull();
        }
    );

    it.each([
        ['S01E02', 1, 2],
        ['s1e12', 1, 12],
        ['S02E123', 2, 123],
        ['1x05', 1, 5],
        ['Show_S01E02_720p', 1, 2],
        ['Ep 3', 1, 3],
        ['Episódio 12', 1, 12],
        ['Cap. 7', 1, 7],
    ])('parses %s', (name, seasonNumber, episodeNumber) => {
        const parsed = parseSeasonEpisode(name);
        expect(parsed?.seasonNumber).toBe(seasonNumber);
        expect(parsed?.episodeNumber).toBe(episodeNumber);
    });

    it('prefers SxxEyy over NxNN when both exist', () => {
        expect(parseSeasonEpisode('Show 3x04 S02E05')).toMatchObject({
            seasonNumber: 2,
            episodeNumber: 5,
            matchedText: 'S02E05',
        });
    });
});
