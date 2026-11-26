import useI18n from '@/components/i18n/provider/i18nContext';
import { formatSize } from '@/shared/utils/formatSize';
import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import LinearProgress from '@mui/material/LinearProgress';
import Typography from '@mui/material/Typography';
import TakeoutImportProgress from './TakeoutImportProgress';
import type { TakeoutQueueEntry } from './takeoutUploadQueue';

type TakeoutQueueItemProps = {
    entry: TakeoutQueueEntry;
    onCancel: (entryId: string) => void;
};

const isCancelable = (entry: TakeoutQueueEntry): boolean =>
    entry.state === 'queued' || entry.state === 'uploading' || entry.state === 'waitingConnection';

const uploadPercentage = (entry: TakeoutQueueEntry): number =>
    entry.file.size === 0
        ? 100
        : Math.min(100, Math.round((entry.uploadedBytes / entry.file.size) * 100));

const TakeoutQueueItem = ({ entry, onCancel }: TakeoutQueueItemProps) => {
    const { t } = useI18n();

    return (
        <div>
            <Typography variant="body2">
                {entry.file.name} ({formatSize(entry.file.size)})
            </Typography>

            {entry.state === 'queued' ? (
                <Typography variant="body2">{t('TAKEOUT_QUEUE_WAITING')}</Typography>
            ) : null}

            {entry.state === 'uploading' ? (
                <>
                    <LinearProgress variant="determinate" value={uploadPercentage(entry)} />
                    <Typography variant="body2">
                        {t('TAKEOUT_UPLOAD_PROGRESS', {
                            sent: formatSize(entry.uploadedBytes),
                            total: formatSize(entry.file.size),
                        })}
                    </Typography>
                </>
            ) : null}

            {entry.state === 'waitingConnection' ? (
                <>
                    <LinearProgress variant="determinate" value={uploadPercentage(entry)} />
                    <Typography variant="body2">
                        {t('TAKEOUT_UPLOAD_WAITING_CONNECTION')}
                    </Typography>
                </>
            ) : null}

            {entry.state === 'completing' ? (
                <>
                    <LinearProgress variant="determinate" value={100} />
                    <Typography variant="body2">{t('TAKEOUT_PROCESSING')}</Typography>
                </>
            ) : null}

            {isCancelable(entry) ? (
                <Button size="small" onClick={() => onCancel(entry.id)}>
                    {t('TAKEOUT_CANCEL')}
                </Button>
            ) : null}

            {entry.state === 'canceled' ? (
                <Typography variant="body2">{t('TAKEOUT_CANCELED')}</Typography>
            ) : null}

            {entry.state === 'canceled' && entry.errorMessage !== '' ? (
                <Alert severity="error">{entry.errorMessage}</Alert>
            ) : null}

            {entry.state === 'done' ? (
                <Alert severity="success">{t('TAKEOUT_UPLOAD_COMPLETE')}</Alert>
            ) : null}

            {entry.state === 'done' && entry.importId !== null ? (
                <TakeoutImportProgress importId={entry.importId} />
            ) : null}

            {entry.state === 'error' || entry.state === 'rejected' ? (
                <Alert severity="error">{entry.errorMessage}</Alert>
            ) : null}
        </div>
    );
};

export default TakeoutQueueItem;
