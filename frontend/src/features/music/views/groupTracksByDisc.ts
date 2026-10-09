import { IMusicData } from '@/features/music/providers/musicProvider/musicProvider';
import { parseDiscNumber } from '@/utils/music';

export type DiscTrackGroup = {
    discNumber: number;
    tracks: Array<{ track: IMusicData; position: number }>;
};

export const groupTracksByDisc = (tracks: IMusicData[]): DiscTrackGroup[] => {
    const discGroups: DiscTrackGroup[] = [];

    tracks.forEach((track, position) => {
        const discNumber = parseDiscNumber(track.metadata);
        const lastGroup = discGroups[discGroups.length - 1];
        if (lastGroup?.discNumber === discNumber) {
            lastGroup.tracks.push({ track, position });
            return;
        }
        discGroups.push({ discNumber, tracks: [{ track, position }] });
    });

    return discGroups;
};
