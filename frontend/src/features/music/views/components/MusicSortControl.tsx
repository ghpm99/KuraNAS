import { Box, IconButton, MenuItem, Select, Tooltip } from '@mui/material';
import { ArrowDown, ArrowUp } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { MusicListSort, MusicListSortField } from '@/types/music';
import { musicSortFieldsByView, type MusicSortableView } from '../musicListSortPreference';

const sortFieldLabelKeys: Record<MusicListSortField, string> = {
    tracks: 'MUSIC_SORT_TRACKS',
    name: 'MUSIC_SORT_NAME',
    recent: 'MUSIC_SORT_RECENT',
    year: 'MUSIC_SORT_YEAR',
};

type MusicSortControlProps = {
    view: MusicSortableView;
    listSort: MusicListSort;
    onFieldChange: (field: MusicListSortField) => void;
    onOrderToggle: () => void;
};

export default function MusicSortControl({
    view,
    listSort,
    onFieldChange,
    onOrderToggle,
}: MusicSortControlProps) {
    const { t } = useI18n();
    const isDescending = listSort.order === 'desc';
    const orderLabel = isDescending ? t('MUSIC_SORT_ORDER_DESC') : t('MUSIC_SORT_ORDER_ASC');

    return (
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, px: 2, pt: 2 }}>
            <Select
                size="small"
                value={listSort.sort}
                onChange={(event) => onFieldChange(event.target.value as MusicListSortField)}
                inputProps={{ 'aria-label': t('MUSIC_SORT_LABEL') }}
            >
                {musicSortFieldsByView[view].map((field) => (
                    <MenuItem key={field} value={field}>
                        {t(sortFieldLabelKeys[field])}
                    </MenuItem>
                ))}
            </Select>
            <Tooltip title={orderLabel}>
                <IconButton size="small" aria-label={orderLabel} onClick={onOrderToggle}>
                    {isDescending ? <ArrowDown size={18} /> : <ArrowUp size={18} />}
                </IconButton>
            </Tooltip>
        </Box>
    );
}
