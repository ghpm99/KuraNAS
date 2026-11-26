import useI18n from '@/components/i18n/provider/i18nContext';
import PageContainer from '@/components/layout/PageContainer';
import PageHeader from '@/components/layout/PageHeader';
import Button from '@mui/material/Button';
import Link from '@mui/material/Link';
import Typography from '@mui/material/Typography';
import useConfirmLeaveWhileUploading from '@/features/files/upload/useConfirmLeaveWhileUploading';
import TakeoutQueueItem from './TakeoutQueueItem';
import TakeoutDropZone from './TakeoutDropZone';
import useTakeoutUploadQueue from './useTakeoutUploadQueue';
import styles from './TakeoutImportScreen.module.css';

const TakeoutImportScreen = () => {
    const { t } = useI18n();
    const { entries, addFile, startQueue, clearSettled, isBusy, hasQueuedEntries } =
        useTakeoutUploadQueue();

    useConfirmLeaveWhileUploading(isBusy);

    return (
        <PageContainer className={styles.content}>
            <PageHeader title={t('TAKEOUT_PAGE_TITLE')} subtitle={t('TAKEOUT_PAGE_DESCRIPTION')} />

            <div className={styles.guide}>
                <Typography variant="subtitle2">{t('TAKEOUT_GUIDE_TITLE')}</Typography>
                <ol className={styles.guideSteps}>
                    <li>
                        <Typography variant="body2" component="span">
                            {t('TAKEOUT_GUIDE_STEP_OPEN')}{' '}
                            <Link
                                href="https://takeout.google.com"
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                {t('TAKEOUT_GUIDE_LINK')}
                            </Link>
                        </Typography>
                    </li>
                    <li>
                        <Typography variant="body2" component="span">
                            {t('TAKEOUT_GUIDE_STEP_SELECT')}
                        </Typography>
                    </li>
                    <li>
                        <Typography variant="body2" component="span">
                            {t('TAKEOUT_GUIDE_STEP_FORMAT')}
                        </Typography>
                    </li>
                    <li>
                        <Typography variant="body2" component="span">
                            {t('TAKEOUT_GUIDE_STEP_DROP')}
                        </Typography>
                    </li>
                </ol>
            </div>

            <TakeoutDropZone onSelectFile={addFile} />

            {entries.map((entry) => (
                <TakeoutQueueItem key={entry.id} entry={entry} />
            ))}

            <div className={styles.actions}>
                <Button
                    variant="contained"
                    onClick={() => void startQueue()}
                    disabled={isBusy || !hasQueuedEntries}
                >
                    {t('TAKEOUT_START_UPLOAD')}
                </Button>
                <Button variant="outlined" onClick={clearSettled}>
                    {t('TAKEOUT_CLEAR')}
                </Button>
            </div>
        </PageContainer>
    );
};

export default TakeoutImportScreen;
