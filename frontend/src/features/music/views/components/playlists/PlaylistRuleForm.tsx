import { Box, Checkbox, FormControlLabel, Stack, TextField, Typography } from '@mui/material';
import ToggleButton from '@mui/material/ToggleButton';
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup';
import { useQuery } from '@tanstack/react-query';
import useI18n from '@/components/i18n/provider/i18nContext';
import useDebouncedValue from '@/components/hooks/useDebouncedValue/useDebouncedValue';
import { previewPlaylistRule } from '@/service/playlist';
import { PlaylistRules } from '@/types/playlist';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import PlaylistRuleOptionsField from './PlaylistRuleOptionsField';
import {
    loadAlbumOptions,
    loadArtistOptions,
    loadFolderOptions,
    loadGenreOptions,
    useRuleOptions,
} from './usePlaylistRuleOptions';
import { hasPlaylistRuleCriteria } from '@/features/music/providers/playlistsProvider/playlistRuleCriteria';

const PREVIEW_DEBOUNCE_MS = 300;

type PlaylistRuleFormProps = {
    rules: PlaylistRules;
    errorField?: string | null;
    onChange: (rules: PlaylistRules) => void;
};

const parseYear = (text: string): number | null => {
    const year = Number(text);
    return text.trim() === '' || !Number.isInteger(year) ? null : year;
};

export default function PlaylistRuleForm({ rules, errorField, onChange }: PlaylistRuleFormProps) {
    const { t } = useI18n();
    const genreOptions = useRuleOptions('genres', loadGenreOptions);
    const artistOptions = useRuleOptions('artists', loadArtistOptions);
    const albumOptions = useRuleOptions('albums', loadAlbumOptions);
    const folderOptions = useRuleOptions('folders', loadFolderOptions);

    const debouncedRules = useDebouncedValue(rules, PREVIEW_DEBOUNCE_MS);
    const hasCriteria = hasPlaylistRuleCriteria(rules);
    const preview = useQuery({
        queryKey: ['playlist-rule-preview', debouncedRules],
        queryFn: () => previewPlaylistRule(debouncedRules),
        enabled: hasPlaylistRuleCriteria(debouncedRules),
        retry: false,
    });

    const isPreviewPending = hasCriteria && (debouncedRules !== rules || preview.isFetching);
    const isRulesFieldInvalid = errorField === 'rules';
    const isYearFieldInvalid = errorField === 'year_from' || errorField === 'year_to';
    const previewErrorMessage = extractBackendErrorMessage(preview.error);

    const renderPreview = () => {
        if (!hasCriteria) {
            return t('MUSIC_PLAYLIST_RULE_PREVIEW_IDLE');
        }
        if (isPreviewPending) {
            return t('MUSIC_PLAYLIST_RULE_PREVIEW_LOADING');
        }
        if (preview.isError) {
            return previewErrorMessage ?? t('MUSIC_PLAYLIST_RULE_PREVIEW_FAILED');
        }
        if (preview.data) {
            return t('MUSIC_PLAYLIST_RULE_PREVIEW_COUNT', {
                count: String(preview.data.track_count),
            });
        }
        return t('MUSIC_PLAYLIST_RULE_PREVIEW_LOADING');
    };

    return (
        <Stack spacing={2} sx={{ mt: 2 }}>
            <ToggleButtonGroup
                exclusive
                size="small"
                value={rules.match}
                onChange={(_event, nextMatch) =>
                    nextMatch && onChange({ ...rules, match: nextMatch })
                }
            >
                <ToggleButton value="any">{t('MUSIC_PLAYLIST_RULE_MATCH_ANY')}</ToggleButton>
                <ToggleButton value="all">{t('MUSIC_PLAYLIST_RULE_MATCH_ALL')}</ToggleButton>
            </ToggleButtonGroup>
            <PlaylistRuleOptionsField
                label={t('MUSIC_PLAYLIST_RULE_GENRES')}
                selectedValues={rules.genres}
                options={genreOptions.data ?? []}
                isLoading={genreOptions.isLoading}
                hasError={isRulesFieldInvalid}
                onChange={(genres) => onChange({ ...rules, genres })}
            />
            <PlaylistRuleOptionsField
                label={t('MUSIC_PLAYLIST_RULE_ARTISTS')}
                selectedValues={rules.artists}
                options={artistOptions.data ?? []}
                isLoading={artistOptions.isLoading}
                hasError={isRulesFieldInvalid}
                onChange={(artists) => onChange({ ...rules, artists })}
            />
            <PlaylistRuleOptionsField
                label={t('MUSIC_PLAYLIST_RULE_ALBUMS')}
                selectedValues={rules.albums}
                options={albumOptions.data ?? []}
                isLoading={albumOptions.isLoading}
                hasError={isRulesFieldInvalid}
                onChange={(albums) => onChange({ ...rules, albums })}
            />
            <PlaylistRuleOptionsField
                label={t('MUSIC_PLAYLIST_RULE_FOLDERS')}
                selectedValues={rules.folders}
                options={folderOptions.data ?? []}
                isLoading={folderOptions.isLoading}
                hasError={isRulesFieldInvalid}
                onChange={(folders) => onChange({ ...rules, folders })}
            />
            <Box sx={{ display: 'flex', gap: 2 }}>
                <TextField
                    fullWidth
                    size="small"
                    type="number"
                    label={t('MUSIC_PLAYLIST_RULE_YEAR_FROM')}
                    value={rules.year_from ?? ''}
                    error={isYearFieldInvalid}
                    onChange={(event) =>
                        onChange({ ...rules, year_from: parseYear(event.target.value) })
                    }
                />
                <TextField
                    fullWidth
                    size="small"
                    type="number"
                    label={t('MUSIC_PLAYLIST_RULE_YEAR_TO')}
                    value={rules.year_to ?? ''}
                    error={isYearFieldInvalid}
                    onChange={(event) =>
                        onChange({ ...rules, year_to: parseYear(event.target.value) })
                    }
                />
            </Box>
            <FormControlLabel
                label={t('MUSIC_PLAYLIST_RULE_FAVORITES_ONLY')}
                control={
                    <Checkbox
                        checked={rules.favorites_only}
                        onChange={(event) =>
                            onChange({ ...rules, favorites_only: event.target.checked })
                        }
                    />
                }
            />
            <Typography
                variant="body2"
                color={preview.isError ? 'error' : 'text.secondary'}
                role="status"
            >
                {renderPreview()}
            </Typography>
        </Stack>
    );
}
