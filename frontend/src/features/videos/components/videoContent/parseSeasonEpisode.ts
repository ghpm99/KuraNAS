export type SeasonEpisodeMatch = {
    seasonNumber: number;
    episodeNumber: number;
    matchedText: string;
};

const tokenStart = '(?:^|[^\\p{L}\\p{N}])';
const tokenEnd = '(?=$|[^\\p{L}\\p{N}])';

const seasonEpisodePattern = new RegExp(
    `${tokenStart}(s(\\d{1,2})\\s?e(\\d{1,3}))${tokenEnd}`,
    'iu'
);
const crossEpisodePattern = new RegExp(`${tokenStart}((\\d{1,2})x(\\d{1,3}))${tokenEnd}`, 'iu');
const keywordEpisodePattern = new RegExp(
    `${tokenStart}((?:ep|episode|epis[oó]dio|cap|cap[ií]tulo)\\.?\\s?(\\d+))${tokenEnd}`,
    'iu'
);

const defaultSeasonNumber = 1;

const toSeasonEpisodeMatch = (
    match: RegExpMatchArray,
    seasonNumber: number,
    episodeNumber: number
): SeasonEpisodeMatch => ({
    seasonNumber,
    episodeNumber,
    matchedText: match[1] ?? '',
});

export const parseSeasonEpisode = (name: string): SeasonEpisodeMatch | null => {
    const seasonEpisodeMatch =
        name.match(seasonEpisodePattern) ?? name.match(crossEpisodePattern);
    if (seasonEpisodeMatch) {
        return toSeasonEpisodeMatch(
            seasonEpisodeMatch,
            Number(seasonEpisodeMatch[2]),
            Number(seasonEpisodeMatch[3])
        );
    }

    const keywordMatch = name.match(keywordEpisodePattern);
    if (keywordMatch) {
        return toSeasonEpisodeMatch(keywordMatch, defaultSeasonNumber, Number(keywordMatch[2]));
    }

    return null;
};
