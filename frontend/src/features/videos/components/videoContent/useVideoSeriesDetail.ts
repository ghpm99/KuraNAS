import { useMemo, useState } from 'react';
import {
    useVideoSeasonEpisodes,
    useVideoSeriesHeader,
} from '@/features/videos/providers/videoContentProvider/useVideoQueries';
import type { VideoSeriesSeasonDto } from '@/service/videoPlayback';

export const VIDEO_SERIES_EPISODES_PAGE_SIZE = 50;

const resolveSeason = (
    seasons: VideoSeriesSeasonDto[],
    chosenSeason: number | null,
    resumeSeason: number | null | undefined
): number | undefined => {
    const availableSeasons = seasons.map((season) => season.season);
    if (chosenSeason !== null && availableSeasons.includes(chosenSeason)) {
        return chosenSeason;
    }
    if (resumeSeason != null && availableSeasons.includes(resumeSeason)) {
        return resumeSeason;
    }
    return availableSeasons[0];
};

export const useVideoSeriesDetail = (playlistId: number) => {
    const [chosenSeason, setChosenSeason] = useState<number | null>(null);
    const headerQuery = useVideoSeriesHeader(playlistId);
    const header = headerQuery.data;
    const seasons = useMemo(() => header?.seasons ?? [], [header?.seasons]);
    const resume = header?.resume ?? null;
    const selectedSeason = resolveSeason(seasons, chosenSeason, resume?.season);

    const episodesQuery = useVideoSeasonEpisodes(
        playlistId,
        selectedSeason,
        VIDEO_SERIES_EPISODES_PAGE_SIZE,
        headerQuery.isSuccess
    );
    const episodes = useMemo(() => {
        const loadedEpisodes = episodesQuery.data?.pages.flatMap((page) => page?.items ?? []) ?? [];
        const episodesById = new Map(loadedEpisodes.map((episode) => [episode.id, episode]));
        return loadedEpisodes.filter((episode) => episodesById.get(episode.id) === episode);
    }, [episodesQuery.data]);

    const totalEpisodes = seasons.reduce((sum, season) => sum + season.episode_count, 0);
    const watchedEpisodes = seasons.reduce((sum, season) => sum + season.watched_count, 0);
    const selectedSeasonEpisodeCount =
        seasons.find((season) => season.season === selectedSeason)?.episode_count ?? 0;

    return {
        headerQuery,
        header,
        seasons,
        resume,
        selectedSeason,
        selectSeason: setChosenSeason,
        episodesQuery,
        episodes,
        totalEpisodes,
        watchedEpisodes,
        selectedSeasonEpisodeCount,
    };
};
