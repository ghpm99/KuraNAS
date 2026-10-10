import { MenuItem, Select } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { VideoMovieSort } from '@/service/videoPlayback';

type VideoMovieSortSelectProps = {
    sort: VideoMovieSort;
    onChange: (sort: VideoMovieSort) => void;
};

const sortLabelKeys: Record<VideoMovieSort, string> = {
    name: 'VIDEO_MOVIES_SORT_NAME',
    recent: 'VIDEO_MOVIES_SORT_RECENT',
};

const sortOptions = Object.keys(sortLabelKeys) as VideoMovieSort[];

export default function VideoMovieSortSelect({ sort, onChange }: VideoMovieSortSelectProps) {
    const { t } = useI18n();

    return (
        <Select
            size="small"
            value={sort}
            onChange={(event) => onChange(event.target.value as VideoMovieSort)}
            inputProps={{ 'aria-label': t('VIDEO_MOVIES_SORT_LABEL') }}
        >
            {sortOptions.map((sortOption) => (
                <MenuItem key={sortOption} value={sortOption}>
                    {t(sortLabelKeys[sortOption])}
                </MenuItem>
            ))}
        </Select>
    );
}
