import useI18n from '@/components/i18n/provider/i18nContext';
import { getTrashSummary } from '@/service/trash';
import { trashQueryKey } from '@/shared/queryKeys/fileQueryKeys';
import { formatSize } from '@/shared/utils/formatSize';
import type { TrashSummary } from '@/types/trash';
import { useQuery } from '@tanstack/react-query';
import styles from './TrashScreen.module.css';

const TrashSummaryLine = () => {
    const { t } = useI18n();
    const summaryQuery = useQuery<TrashSummary>({
        queryKey: [trashQueryKey, 'summary'],
        queryFn: getTrashSummary,
        retry: false,
    });

    const itemCount = summaryQuery.data?.items;
    if (itemCount === undefined || itemCount <= 0) {
        return null;
    }

    const occupiedSize = formatSize(summaryQuery.data?.size_bytes ?? 0);
    const summaryText =
        itemCount === 1
            ? t('TRASH_SUMMARY_ONE', { size: occupiedSize })
            : t('TRASH_SUMMARY', { count: String(itemCount), size: occupiedSize });

    return <p className={styles.summary}>{summaryText}</p>;
};

export default TrashSummaryLine;
