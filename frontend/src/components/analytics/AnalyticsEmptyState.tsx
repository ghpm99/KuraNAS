import { Typography } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';

export default function AnalyticsEmptyState({ messageKey }: { messageKey: string }) {
    const { t } = useI18n();
    return <Typography color="text.secondary">{t(messageKey)}</Typography>;
}
