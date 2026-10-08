import useI18n from '@/components/i18n/provider/i18nContext';
import { Tooltip } from '@mui/material';
import { Snowflake } from 'lucide-react';

const ColdTierIndicator = ({ size = 14 }: { size?: number }) => {
    const { t } = useI18n();
    const label = t('FILE_TIER_COLD_INDICATOR');

    return (
        <Tooltip title={label}>
            <span
                role="img"
                aria-label={label}
                style={{ display: 'inline-flex', alignItems: 'center', color: '#60a5fa' }}
            >
                <Snowflake size={size} />
            </span>
        </Tooltip>
    );
};

export default ColdTierIndicator;
