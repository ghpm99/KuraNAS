import { Alert, Box, CircularProgress, Typography } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import TextFileViewer from '@/features/files/fileContent/components/fileViewer/textFileViewer';
import useTextFilePreview from '@/features/files/fileContent/components/fileViewer/useTextFilePreview';
import { textPreviewMaxBytes } from '@/service/files';
import { formatSize } from '@/shared/utils/formatSize';
import DocumentCsvView from './DocumentCsvView';
import DocumentMarkdownView from './DocumentMarkdownView';

type DocumentFormattedTextProps = {
    fileId: number;
    fileSize: number;
    kind: 'plainText' | 'markdown' | 'csv';
};

const DocumentFormattedText = ({ fileId, fileSize, kind }: DocumentFormattedTextProps) => {
    const { t } = useI18n();
    const { data: preview, isPending, isError } = useTextFilePreview(fileId, fileSize);

    if (kind === 'plainText') return <TextFileViewer file={{ id: fileId, size: fileSize }} />;

    if (isPending) {
        return (
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, p: 2 }}>
                <CircularProgress size={16} />
                <Typography variant="body2">{t('TEXT_VIEWER_LOADING')}</Typography>
            </Box>
        );
    }

    if (isError || !preview) return <Alert severity="warning">{t('TEXT_VIEWER_ERROR')}</Alert>;

    return (
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1, minWidth: 0 }}>
            {preview.isTruncated ? (
                <Alert severity="info">
                    {t('TEXT_VIEWER_TRUNCATED', { size: formatSize(textPreviewMaxBytes) })}
                </Alert>
            ) : null}
            {kind === 'markdown' ? (
                <DocumentMarkdownView source={preview.text} />
            ) : (
                <DocumentCsvView source={preview.text} />
            )}
        </Box>
    );
};

export default DocumentFormattedText;
