import Message from '@/components/ui/Message/Message';
import useI18n from '@/components/i18n/provider/i18nContext';

export default function AnalyticsErrorState({ messageKey }: { messageKey: string }) {
	const { t } = useI18n();
	return <Message type='error' text={t(messageKey)} />;
}
