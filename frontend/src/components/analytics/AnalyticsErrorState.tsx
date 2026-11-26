import ErrorState from '@/components/errorState/errorState';
import Message from '@/components/ui/Message/Message';
import useI18n from '@/components/i18n/provider/i18nContext';

interface AnalyticsErrorStateProps {
    messageKey: string;
    backendMessage?: string;
    onRetry?: () => void;
}

export default function AnalyticsErrorState({
    messageKey,
    backendMessage,
    onRetry,
}: AnalyticsErrorStateProps) {
    const { t } = useI18n();

    if (!onRetry) return <Message type="error" text={t(messageKey)} />;

    return <ErrorState title={t(messageKey)} backendMessage={backendMessage} onRetry={onRetry} />;
}
