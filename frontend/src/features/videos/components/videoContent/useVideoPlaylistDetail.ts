import { useMemo } from 'react';
import { parseSeasonEpisode } from './parseSeasonEpisode';
import { type VideoPlaylistDto, type VideoPlaylistItemDto } from '@/service/videoPlayback';

export type VideoDetailItem = VideoPlaylistItemDto & {
    displayTitle: string;
    sequenceLabel: string;
    seasonNumber: number | null;
    episodeNumber: number | null;
};

export type VideoSeasonGroup = {
    key: string;
    label: string;
    items: VideoDetailItem[];
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

const parseEpisode = (name: string): ParsedEpisode => {
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

export const buildVideoPlaylistDetail = (playlist: VideoPlaylistDto) => {
    const orderedItems = [...playlist.items]
        .sort((a, b) => a.order_index - b.order_index || a.id - b.id)
        .map((item) => {
            const parsed = parseEpisode(item.video.name);
            return {
                ...item,
                ...parsed,
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

    const groupedSeasons = hasEpisodeData
        ? orderedItems.reduce<VideoSeasonGroup[]>((groups, item) => {
              const seasonNumber = item.seasonNumber ?? 1;
              const key = `season-${seasonNumber}`;
              const existingGroup = groups.find((group) => group.key === key);
              if (existingGroup) {
                  existingGroup.items.push(item);
                  return groups;
              }

              return [
                  ...groups,
                  {
                      key,
                      label: String(seasonNumber),
                      items: [item],
                  },
              ];
          }, [])
        : [];

    return {
        orderedItems,
        groupedSeasons,
        completedCount,
        hasEpisodeData,
        resumeItem,
    };
};

export const useVideoPlaylistDetail = (playlist: VideoPlaylistDto) => {
    return useMemo(() => buildVideoPlaylistDetail(playlist), [playlist]);
};
