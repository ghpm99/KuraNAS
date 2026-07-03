import useI18n from '@/components/i18n/provider/i18nContext';
import PageContainer from '@/components/layout/PageContainer';
import PageHeader from '@/components/layout/PageHeader';
import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import LinearProgress from '@mui/material/LinearProgress';
import Typography from '@mui/material/Typography';
import TakeoutDropZone from './TakeoutDropZone';
import useTakeoutUpload from './useTakeoutUpload';
import styles from './TakeoutImportScreen.module.css';

const TakeoutImportScreen = () => {
    const { t } = useI18n();
    const {
        state,
        progress,
        fileName,
        jobId,
        errorMessage,
        progressMessage,
        selectFile,
        startUpload,
        reset,
    } = useTakeoutUpload();

    return (
        <PageContainer className={styles.content}>
            <PageHeader title={t('TAKEOUT_PAGE_TITLE')} subtitle={t('TAKEOUT_PAGE_DESCRIPTION')} />

            <TakeoutDropZone onSelectFile={selectFile} />

            {fileName ? <Typography variant="body2">{fileName}</Typography> : null}

            {state === 'uploading' || state === 'completing' ? (
                <>
                    <LinearProgress variant="determinate" value={progress} />
                    <Typography variant="body2">
                        {state === 'completing' ? t('TAKEOUT_PROCESSING') : progressMessage}
                    </Typography>
                </>
            ) : null}

            {state === 'done' ? (
                <Alert severity="success">
                    {t('TAKEOUT_UPLOAD_COMPLETE')} {jobId ? `(job #${jobId})` : ''}
                </Alert>
            ) : null}

            {state === 'error' ? <Alert severity="error">{errorMessage}</Alert> : null}

            <div className={styles.actions}>
                <Button
                    variant="contained"
                    onClick={() => void startUpload()}
                    disabled={state === 'uploading' || state === 'completing' || state === 'idle'}
                >
                    {t('TAKEOUT_UPLOADING')}
                </Button>
                <Button variant="outlined" onClick={reset}>
                    {t('SETTINGS_RESET')}
                </Button>
            </div>
        </PageContainer>
    );
};

export default TakeoutImportScreen;
