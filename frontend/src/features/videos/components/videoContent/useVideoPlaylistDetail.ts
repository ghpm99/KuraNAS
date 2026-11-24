import { useMemo } from 'react';
import { parseSeasonEpisode } from './parseSeasonEpisode';
import { type VideoPlaylistDto, type VideoPlaylistItemDto } from '@/service/videoPlayback';

export type VideoDetailItem = VideoPlaylistItemDto & {
    displayTitle: string;
    sequenceLabel: string;
    seasonNumber: number | null;
    episodeNumber: number | null;
};

type ParsedEpisode = {
    displayTitle: string;
    seasonNumber: number | null;
    episodeNumber: number | null;
    sequenceLabel: string;
};

const extensionPattern = /\.[^/.]+$/;
const whitespacePattern = /[._-]+/g;
const spaceCollapsePattern = /\s+/g;

export const parseEpisode = (name: string): ParsedEpisode => {
    const cleanName = name.replace(extensionPattern, '');

    const seasonEpisode = parseSeasonEpisode(cleanName);
    if (seasonEpisode) {
        const { seasonNumber, episodeNumber, matchedText } = seasonEpisode;
        const displayTitle = cleanName
            .replace(matchedText, ' ')
            .replace(whitespacePattern, ' ')
            .replace(spaceCollapsePattern, ' ')
            .trim();
        const isValidEpisode = episodeNumber > 0;

        return {
            displayTitle: displayTitle || cleanName,
            seasonNumber,
            episodeNumber: isValidEpisode ? episodeNumber : null,
            sequenceLabel: isValidEpisode
                ? `S${String(seasonNumber).padStart(2, '0')}E${String(episodeNumber).padStart(2, '0')}`
                : '',
        };
    }

    return {
        displayTitle:
            cleanName.replace(whitespacePattern, ' ').replace(spaceCollapsePattern, ' ').trim() ||
            cleanName,
        seasonNumber: null,
        episodeNumber: null,
        sequenceLabel: '',
    };
};

const buildSequenceLabel = (seasonNumber: number | null, episodeNumber: number | null): string =>
    episodeNumber === null || episodeNumber <= 0
        ? ''
        : `S${String(seasonNumber ?? 1).padStart(2, '0')}E${String(episodeNumber).padStart(2, '0')}`;

export const buildVideoPlaylistDetail = (playlist: VideoPlaylistDto) => {
    const orderedItems = [...playlist.items]
        .sort((a, b) => a.order_index - b.order_index || a.id - b.id)
        .map((item) => {
            const parsed = parseEpisode(item.video.name);
            const seasonNumber = item.season ?? parsed.seasonNumber;
            const episodeNumber = item.episode ?? parsed.episodeNumber;
            return {
                ...item,
                displayTitle: parsed.displayTitle,
                seasonNumber,
                episodeNumber,
                sequenceLabel: buildSequenceLabel(seasonNumber, episodeNumber),
            };
        });

    const completedCount = orderedItems.filter((item) => item.status === 'completed').length;
    const inProgressItem = orderedItems.find((item) => item.status === 'in_progress') ?? null;
    const resumeItem =
        inProgressItem ??
        orderedItems.find((item) => item.status !== 'completed') ??
        orderedItems[0] ??
        null;
    const hasEpisodeData = orderedItems.some((item) => item.episodeNumber !== null);

    return {
        orderedItems,
        completedCount,
        hasEpisodeData,
        resumeItem,
    };
};

export const useVideoPlaylistDetail = (playlist: VideoPlaylistDto) => {
    return useMemo(() => buildVideoPlaylistDetail(playlist), [playlist]);
};
