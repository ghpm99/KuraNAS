import { useState } from 'react';
import { Button, IconButton, LinearProgress, Tooltip } from '@mui/material';
import { ChevronDown, ChevronUp, RotateCcw, X } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import ConflictPolicySelect from './conflictPolicySelect';
import useUploadQueueContext from './uploadQueueContext';
import { isActiveItem, isRetryableItem, type UploadItem } from './uploadQueueTypes';
import styles from './uploadQueuePanel.module.css';

const statusLabelKeys: Record<UploadItem['status'], string> = {
    queued: 'FILES_UPLOAD_STATUS_QUEUED',
    uploading: 'FILES_UPLOAD_STATUS_UPLOADING',
    done: 'FILES_UPLOAD_STATUS_DONE',
    renamed: 'FILES_UPLOAD_STATUS_RENAMED',
    replaced: 'FILES_UPLOAD_STATUS_REPLACED',
    skipped: 'FILES_UPLOAD_STATUS_SKIPPED',
    failed: 'FILES_UPLOAD_STATUS_FAILED',
    canceled: 'FILES_UPLOAD_STATUS_CANCELED',
};

const UploadQueuePanel = () => {
    const { t } = useI18n();
    const { items, cancel, cancelAll, retry, retryFailed, clearFinished } = useUploadQueueContext();
    const [isCollapsed, setIsCollapsed] = useState(false);

    if (items.length === 0) return null;

    const activeCount = items.filter(isActiveItem).length;
    const finishedCount = items.length - activeCount;
    const hasFailed = items.some((item) => item.status === 'failed');
    const toggleLabel = t(isCollapsed ? 'FILES_UPLOAD_EXPAND' : 'FILES_UPLOAD_COLLAPSE');

    return (
        <aside className={styles.panel} aria-label={t('FILES_UPLOAD_PANEL_TITLE')}>
            <header className={styles.header}>
                <strong className={styles.title} role="status" aria-live="polite">
                    {`${t('FILES_UPLOAD_PANEL_TITLE')} ${finishedCount}/${items.length}`}
                </strong>
                <Tooltip title={toggleLabel}>
                    <IconButton size="small" aria-label={toggleLabel} onClick={() => setIsCollapsed(!isCollapsed)}>
                        {isCollapsed ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
                    </IconButton>
                </Tooltip>
            </header>

            {isCollapsed ? null : (
                <>
                    <div className={styles.toolbar}>
                        <ConflictPolicySelect />
                        <div className={styles.toolbarActions}>
                            {activeCount > 0 ? (
                                <Button size="small" onClick={cancelAll}>
                                    {t('FILES_UPLOAD_CANCEL_ALL')}
                                </Button>
                            ) : null}
                            {hasFailed ? (
                                <Button size="small" onClick={retryFailed}>
                                    {t('FILES_UPLOAD_RETRY_FAILED')}
                                </Button>
                            ) : null}
                            {finishedCount > 0 ? (
                                <Button size="small" onClick={clearFinished}>
                                    {t('FILES_UPLOAD_CLEAR_FINISHED')}
                                </Button>
                            ) : null}
                        </div>
                    </div>

                    <ul className={styles.list}>
                        {items.map((item) => (
                            <li key={item.id} className={styles.row}>
                                <div className={styles.rowMain}>
                                    <span className={styles.fileName} title={item.displayName}>
                                        {item.displayName}
                                    </span>
                                    <LinearProgress
                                        variant="determinate"
                                        value={item.progress}
                                        aria-label={item.displayName}
                                    />
                                    <span
                                        className={
                                            item.status === 'failed'
                                                ? `${styles.status} ${styles.statusFailed}`
                                                : styles.status
                                        }
                                    >
                                        {t(statusLabelKeys[item.status])}
                                        {item.status === 'failed'
                                            ? `: ${item.error || t('ERROR_UPLOAD_FAILED')}`
                                            : ''}
                                        {item.status === 'renamed' && item.savedName
                                            ? `: ${item.savedName}`
                                            : ''}
                                    </span>
                                </div>
                                {isActiveItem(item) ? (
                                    <Tooltip title={t('FILES_UPLOAD_CANCEL')}>
                                        <IconButton
                                            size="small"
                                            aria-label={`${t('FILES_UPLOAD_CANCEL')} ${item.displayName}`}
                                            onClick={() => cancel(item.id)}
                                        >
                                            <X size={16} />
                                        </IconButton>
                                    </Tooltip>
                                ) : null}
                                {isRetryableItem(item) ? (
                                    <Tooltip title={t('FILES_UPLOAD_RETRY')}>
                                        <IconButton
                                            size="small"
                                            aria-label={`${t('FILES_UPLOAD_RETRY')} ${item.displayName}`}
                                            onClick={() => retry(item.id)}
                                        >
                                            <RotateCcw size={16} />
                                        </IconButton>
                                    </Tooltip>
                                ) : null}
                            </li>
                        ))}
                    </ul>
                </>
            )}
        </aside>
    );
};

export default UploadQueuePanel;
