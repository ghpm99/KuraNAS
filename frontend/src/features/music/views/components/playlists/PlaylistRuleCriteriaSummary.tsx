import { Box, Chip, Typography } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import { PlaylistRules } from '@/types/playlist';
import { formatAlbumKeyLabel } from './formatAlbumKeyLabel';
import {
    loadAlbumOptions,
    loadArtistOptions,
    loadGenreOptions,
    useRuleOptions,
} from './usePlaylistRuleOptions';
import { PlaylistRuleOption } from './PlaylistRuleOptionsField';

const labelFromOptions = (
    options: PlaylistRuleOption[] | undefined,
    key: string,
    fallbackLabel: (key: string) => string = (unresolvedKey) => unresolvedKey
): string => options?.find((option) => option.value === key)?.label ?? fallbackLabel(key);

type CriteriaGroup = { label: string; values: string[] };

type PlaylistRuleCriteriaSummaryProps = {
    rules?: PlaylistRules | null;
};

export default function PlaylistRuleCriteriaSummary({ rules }: PlaylistRuleCriteriaSummaryProps) {
    const { t } = useI18n();
    const genreOptions = useRuleOptions('genres', loadGenreOptions);
    const artistOptions = useRuleOptions('artists', loadArtistOptions);
    const albumOptions = useRuleOptions('albums', loadAlbumOptions);
    if (!rules) return null;

    const yearRange = [rules.year_from, rules.year_to].filter((year) => year != null);
    const criteriaGroups: CriteriaGroup[] = [
        {
            label: t('MUSIC_PLAYLIST_RULE_GENRES'),
            values: (rules.genres ?? []).map((key) => labelFromOptions(genreOptions.data, key)),
        },
        {
            label: t('MUSIC_PLAYLIST_RULE_ARTISTS'),
            values: (rules.artists ?? []).map((key) => labelFromOptions(artistOptions.data, key)),
        },
        {
            label: t('MUSIC_PLAYLIST_RULE_ALBUMS'),
            values: (rules.albums ?? []).map((key) =>
                labelFromOptions(albumOptions.data, key, formatAlbumKeyLabel)
            ),
        },
        { label: t('MUSIC_PLAYLIST_RULE_FOLDERS'), values: rules.folders ?? [] },
    ].filter((group) => group.values.length > 0);

    return (
        <Box data-testid="playlist-rule-criteria" sx={{ mb: 2 }}>
            <Typography variant="subtitle2" sx={{ mb: 1 }}>
                {t(
                    rules.match === 'all'
                        ? 'MUSIC_PLAYLIST_RULE_MATCH_ALL'
                        : 'MUSIC_PLAYLIST_RULE_MATCH_ANY'
                )}
            </Typography>
            {criteriaGroups.map((group) => (
                <Box
                    key={group.label}
                    sx={{
                        display: 'flex',
                        flexWrap: 'wrap',
                        alignItems: 'center',
                        gap: 0.5,
                        mb: 0.5,
                    }}
                >
                    <Typography variant="caption" color="text.secondary" sx={{ mr: 0.5 }}>
                        {group.label}
                    </Typography>
                    {group.values.map((value) => (
                        <Chip key={value} size="small" label={value} />
                    ))}
                </Box>
            ))}
            {yearRange.length > 0 && (
                <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5, mb: 0.5 }}>
                    {rules.year_from != null && (
                        <Chip
                            size="small"
                            label={`${t('MUSIC_PLAYLIST_RULE_YEAR_FROM')}: ${rules.year_from}`}
                        />
                    )}
                    {rules.year_to != null && (
                        <Chip
                            size="small"
                            label={`${t('MUSIC_PLAYLIST_RULE_YEAR_TO')}: ${rules.year_to}`}
                        />
                    )}
                </Box>
            )}
            {rules.favorites_only && (
                <Chip size="small" label={t('MUSIC_PLAYLIST_RULE_FAVORITES_ONLY')} />
            )}
            <Typography variant="caption" color="text.secondary" component="div" sx={{ mt: 1 }}>
                {t('MUSIC_PLAYLIST_RULE_REMOVED_NOTE')}
            </Typography>
        </Box>
    );
}
