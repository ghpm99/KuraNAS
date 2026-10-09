import useI18n from '@/components/i18n/provider/i18nContext';
import { Box, Tooltip } from '@mui/material';
import { Snowflake } from 'lucide-react';

const coldIndicatorLightColor = '#1d4ed8';
const coldIndicatorDarkColor = '#60a5fa';

const ColdTierIndicator = ({ size = 14 }: { size?: number }) => {
    const { t } = useI18n();
    const label = t('FILE_TIER_COLD_INDICATOR');

    return (
        <Tooltip title={label}>
            <Box
                component="span"
                role="img"
                aria-label={label}
                sx={{
                    display: 'inline-flex',
                    alignItems: 'center',
                    color: (theme) =>
                        theme.palette.mode === 'dark' ? coldIndicatorDarkColor : coldIndicatorLightColor,
                }}
            >
                <Snowflake size={size} />
            </Box>
        </Tooltip>
    );
};

export default ColdTierIndicator;
