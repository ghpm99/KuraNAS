import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import useImageClassificationBackfill from './useImageClassificationBackfill';

type ImageClassificationBackfillProps = {
    disabled?: boolean;
};

const ImageClassificationBackfill = ({ disabled = false }: ImageClassificationBackfillProps) => {
    const { t, pendingCount, isLoading, hasError, isStarting, startBackfill } =
        useImageClassificationBackfill();

    return (
        <div>
            <p>{t('IMAGE_CLASSIFY_BACKFILL_DESCRIPTION')}</p>

            {isLoading ? <CircularProgress size={20} /> : null}
            {hasError ? (
                <Alert severity="error">{t('IMAGE_CLASSIFY_BACKFILL_TOAST_ERROR')}</Alert>
            ) : null}
            {!isLoading && !hasError ? (
                <Chip
                    variant="outlined"
                    label={
                        pendingCount > 0
                            ? t('IMAGE_CLASSIFY_BACKFILL_PENDING_LABEL', {
                                  count: String(pendingCount),
                              })
                            : t('IMAGE_CLASSIFY_BACKFILL_NONE')
                    }
                />
            ) : null}

            <Button
                variant="contained"
                disabled={disabled || isStarting || isLoading || pendingCount === 0}
                onClick={startBackfill}
                startIcon={isStarting ? <CircularProgress size={16} color="inherit" /> : undefined}
            >
                {t('IMAGE_CLASSIFY_BACKFILL_BUTTON')}
            </Button>
        </div>
    );
};

export default ImageClassificationBackfill;
