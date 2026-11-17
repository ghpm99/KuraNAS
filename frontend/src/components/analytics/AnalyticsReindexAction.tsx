import Button from '@/components/ui/Button/Button';
import useI18n from '@/components/i18n/provider/i18nContext';
import useReindexJob from '@/shared/hooks/useReindexJob';
import { CircularProgress } from '@mui/material';

interface AnalyticsReindexActionProps {
    onFinished: () => void;
}

const AnalyticsReindexAction = ({ onFinished }: AnalyticsReindexActionProps) => {
    const { t } = useI18n();
    const { start, isRunning } = useReindexJob({ onFinished });

    return (
        <Button
            variant="secondary"
            disabled={isRunning}
            onClick={() => void start({ scope: 'all' })}
        >
            {isRunning && (
                <CircularProgress size={16} aria-label={t('ANALYTICS_REINDEX_RUNNING')} />
            )}
            {t('ANALYTICS_ACTION_REINDEX')}
        </Button>
    );
};

export default AnalyticsReindexAction;
