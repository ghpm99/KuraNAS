import useI18n from '@/components/i18n/provider/i18nContext';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import { textPreviewMaxBytes } from '@/service/files';
import { formatSize } from '@/shared/utils/formatSize';
import { Alert, Box, CircularProgress, FormControlLabel, Switch, Typography } from '@mui/material';
import { useState } from 'react';
import useTextFilePreview from './useTextFilePreview';

const TextFileViewer = ({ file }: { file: Pick<FileData, 'id' | 'size'> }) => {
    const { t } = useI18n();
    const [isLineWrapEnabled, setIsLineWrapEnabled] = useState(false);
    const { data: preview, isPending, isError } = useTextFilePreview(file.id, file.size);

    if (isPending) {
        return (
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, p: 2 }}>
                <CircularProgress size={16} />
                <Typography variant="body2">{t('TEXT_VIEWER_LOADING')}</Typography>
            </Box>
        );
    }

    if (isError || !preview) {
        return (
            <Alert severity="warning" sx={{ m: 2 }}>
                {t('TEXT_VIEWER_ERROR')}
            </Alert>
        );
    }

    return (
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1, minWidth: 0, width: '100%' }}>
            <FormControlLabel
                sx={{ alignSelf: 'flex-end', mr: 0 }}
                control={
                    <Switch
                        size="small"
                        checked={isLineWrapEnabled}
                        onChange={(event) => setIsLineWrapEnabled(event.target.checked)}
                    />
                }
                label={t('TEXT_VIEWER_WRAP')}
            />
            {preview.isTruncated ? (
                <Alert severity="info">
                    {t('TEXT_VIEWER_TRUNCATED', { size: formatSize(textPreviewMaxBytes) })}
                </Alert>
            ) : null}
            {preview.text === '' ? (
                <Typography variant="body2" color="text.secondary">
                    {t('TEXT_VIEWER_EMPTY')}
                </Typography>
            ) : (
                <Box
                    component="pre"
                    tabIndex={0}
                    sx={{
                        m: 0,
                        p: 1.5,
                        maxHeight: '70vh',
                        overflow: 'auto',
                        fontFamily: 'monospace',
                        fontSize: '0.85rem',
                        whiteSpace: isLineWrapEnabled ? 'pre-wrap' : 'pre',
                        overflowWrap: isLineWrapEnabled ? 'anywhere' : 'normal',
                        borderRadius: 1,
                        border: '1px solid',
                        borderColor: 'divider',
                    }}
                >
                    {preview.text}
                </Box>
            )}
        </Box>
    );
};

export default TextFileViewer;
