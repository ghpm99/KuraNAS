import useI18n from '@/components/i18n/provider/i18nContext';
import PageContainer from '@/components/layout/PageContainer';
import PageHeader from '@/components/layout/PageHeader';
import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import Link from '@mui/material/Link';
import LinearProgress from '@mui/material/LinearProgress';
import Typography from '@mui/material/Typography';
import useConfirmLeaveWhileUploading from '@/features/files/upload/useConfirmLeaveWhileUploading';
import TakeoutImportProgress from './TakeoutImportProgress';
import TakeoutDropZone from './TakeoutDropZone';
import useTakeoutUpload from './useTakeoutUpload';
import styles from './TakeoutImportScreen.module.css';

const TakeoutImportScreen = () => {
    const { t } = useI18n();
    const {
        state,
        progress,
        fileName,
        importId,
        errorMessage,
        progressMessage,
        selectFile,
        startUpload,
        reset,
    } = useTakeoutUpload();
    const isUploadInProgress = state === 'uploading' || state === 'completing';

    useConfirmLeaveWhileUploading(isUploadInProgress);

    return (
        <PageContainer className={styles.content}>
            <PageHeader title={t('TAKEOUT_PAGE_TITLE')} subtitle={t('TAKEOUT_PAGE_DESCRIPTION')} />

            <div className={styles.guide}>
                <Typography variant="subtitle2">{t('TAKEOUT_GUIDE_TITLE')}</Typography>
                <ol className={styles.guideSteps}>
                    <li>
                        <Typography variant="body2" component="span">
                            {t('TAKEOUT_GUIDE_STEP_OPEN')}{' '}
                            <Link href="https://takeout.google.com" target="_blank" rel="noopener noreferrer">
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

            <TakeoutDropZone onSelectFile={selectFile} />

            {fileName ? <Typography variant="body2">{fileName}</Typography> : null}

            {isUploadInProgress ? (
                <>
                    <LinearProgress variant="determinate" value={progress} />
                    <Typography variant="body2">
                        {state === 'completing' ? t('TAKEOUT_PROCESSING') : progressMessage}
                    </Typography>
                </>
            ) : null}

            {state === 'done' ? (
                <Alert severity="success">{t('TAKEOUT_UPLOAD_COMPLETE')}</Alert>
            ) : null}

            {state === 'done' && importId !== null ? <TakeoutImportProgress importId={importId} /> : null}

            {state === 'error' ? <Alert severity="error">{errorMessage}</Alert> : null}

            <div className={styles.actions}>
                <Button
                    variant="contained"
                    onClick={() => void startUpload()}
                    disabled={isUploadInProgress || state === 'idle'}
                >
                    {t('TAKEOUT_START_UPLOAD')}
                </Button>
                <Button variant="outlined" onClick={reset}>
                    {t('TAKEOUT_CLEAR')}
                </Button>
            </div>
        </PageContainer>
    );
};

export default TakeoutImportScreen;
