import { MenuItem, TextField } from '@mui/material';
import { useQuery } from '@tanstack/react-query';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getCapturePlatforms } from '@/service/captures';

type CapturePlatformFilterProps = {
    selectedPlatform: string;
    onChange: (platform: string) => void;
};

const withSelectedPlatform = (platforms: string[], selectedPlatform: string): string[] =>
    selectedPlatform && !platforms.includes(selectedPlatform)
        ? [...platforms, selectedPlatform]
        : platforms;

const CapturePlatformFilter = ({ selectedPlatform, onChange }: CapturePlatformFilterProps) => {
    const { t } = useI18n();
    const platformsQuery = useQuery({
        queryKey: ['capture-platforms'],
        queryFn: getCapturePlatforms,
    });
    const platforms = withSelectedPlatform(platformsQuery.data ?? [], selectedPlatform);

    if (platforms.length === 0) return null;

    return (
        <TextField
            select
            size="small"
            label={t('CAPTURES_FILTER_PLATFORM')}
            value={selectedPlatform}
            onChange={(event) => onChange(event.target.value)}
            sx={{ minWidth: 180 }}
        >
            <MenuItem value="">{t('CAPTURES_FILTER_PLATFORM_ALL')}</MenuItem>
            {platforms.map((platform) => (
                <MenuItem key={platform} value={platform}>
                    {platform}
                </MenuItem>
            ))}
        </TextField>
    );
};

export default CapturePlatformFilter;
