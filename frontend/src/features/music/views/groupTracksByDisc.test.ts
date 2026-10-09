import { IMusicData } from '@/features/music/providers/musicProvider/musicProvider';
import { groupTracksByDisc } from './groupTracksByDisc';

const trackOnDisc = (id: number, discNumber?: string) =>
    ({ id, name: `t${id}`, metadata: { disc_number: discNumber } }) as unknown as IMusicData;

describe('groupTracksByDisc', () => {
    it('returns no groups for an empty list', () => {
        expect(groupTracksByDisc([])).toEqual([]);
    });

    it('puts tracks without disc information on disc 1', () => {
        const groups = groupTracksByDisc([trackOnDisc(1), trackOnDisc(2, '')]);

        expect(groups).toHaveLength(1);
        expect(groups[0]?.discNumber).toBe(1);
        expect(groups[0]?.tracks.map((entry) => entry.position)).toEqual([0, 1]);
    });

    it('splits consecutive tracks by disc keeping the list position', () => {
        const groups = groupTracksByDisc([
            trackOnDisc(1, '1/2'),
            trackOnDisc(2, '1/2'),
            trackOnDisc(3, '2/2'),
        ]);

        expect(groups.map((group) => group.discNumber)).toEqual([1, 2]);
        expect(groups[1]?.tracks.map((entry) => entry.track.id)).toEqual([3]);
        expect(groups[1]?.tracks[0]?.position).toBe(2);
    });
});
