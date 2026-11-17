import type { IMusicData } from '@/types/music';
import { readInitialFormValues } from './trackMetadataFormValues';

const trackWithGenre = (genre: string | undefined): IMusicData =>
    ({ id: 1, name: 'one.mp3', metadata: { genre } }) as IMusicData;

describe('readInitialFormValues genres', () => {
    it('splits the genre tag on semicolon, comma and pipe', () => {
        expect(readInitialFormValues(trackWithGenre('Rock; Pop, Funk | Jazz')).genres).toEqual([
            'Rock',
            'Pop',
            'Funk',
            'Jazz',
        ]);
    });

    it('keeps a slash inside the genre so the backend decides how to split it', () => {
        expect(readInitialFormValues(trackWithGenre('Rock/Pop')).genres).toEqual(['Rock/Pop']);
        expect(readInitialFormValues(trackWithGenre('R&B/Soul; Pop')).genres).toEqual([
            'R&B/Soul',
            'Pop',
        ]);
    });

    it('returns no genres when the tag is missing or blank', () => {
        expect(readInitialFormValues(trackWithGenre(undefined)).genres).toEqual([]);
        expect(readInitialFormValues(trackWithGenre(' ; ')).genres).toEqual([]);
    });
});
