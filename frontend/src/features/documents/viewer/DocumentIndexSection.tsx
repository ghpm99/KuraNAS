import { Alert, Box, Button, Typography } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import { formatSize } from '@/shared/utils/formatSize';
import DocumentViewerLoading from './DocumentViewerLoading';
import useDocumentIndexState from './useDocumentIndexState';
import useReindexDocument from './useReindexDocument';

const reasonMessageKeys: Record<string, string> = {
    encrypted: 'DOCUMENTS_INDEX_REASON_ENCRYPTED',
    no_text: 'DOCUMENTS_INDEX_REASON_NO_TEXT',
    too_large: 'DOCUMENTS_INDEX_REASON_TOO_LARGE',
    unsupported_format: 'DOCUMENTS_INDEX_REASON_UNSUPPORTED_FORMAT',
    binary: 'DOCUMENTS_INDEX_REASON_BINARY',
    timeout: 'DOCUMENTS_INDEX_REASON_TIMEOUT',
    unreadable: 'DOCUMENTS_INDEX_REASON_UNREADABLE',
};

const DocumentIndexSection = ({ fileId }: { fileId: number }) => {
    const { t } = useI18n();
    const { data: indexState, isPending, isError } = useDocumentIndexState(fileId);
    const reindex = useReindexDocument(fileId);
    const reasonMessageKey = reasonMessageKeys[indexState?.reason ?? ''];

    return (
        <Box sx={{ px: 2, pb: 2, display: 'flex', flexDirection: 'column', gap: 1 }}>
            <Typography component="h2" variant="subtitle2" fontWeight={600}>
                {t('DOCUMENTS_INDEX_SECTION_TITLE')}
            </Typography>
            {isPending ? <DocumentViewerLoading label={t('DOCUMENTS_INDEX_LOADING')} /> : null}
            {isError ? <Alert severity="warning">{t('DOCUMENTS_INDEX_LOAD_ERROR')}</Alert> : null}
            {indexState ? (
                <Box component="dl" sx={{ m: 0, display: 'grid', gap: 0.5 }}>
                    <Typography component="dt" variant="caption" color="text.secondary">
                        {t('DOCUMENTS_INDEX_LABEL_STATE')}
                    </Typography>
                    <Typography component="dd" variant="body2" sx={{ m: 0 }}>
                        {t(`DOCUMENTS_INDEX_STATE_${String(indexState.state).toUpperCase()}`)}
                    </Typography>
                    <Typography component="dt" variant="caption" color="text.secondary">
                        {t('DOCUMENTS_INDEX_LABEL_BYTES')}
                    </Typography>
                    <Typography component="dd" variant="body2" sx={{ m: 0 }}>
                        {formatSize(indexState.indexed_bytes ?? 0)}
                    </Typography>
                    {reasonMessageKey ? (
                        <>
                            <Typography component="dt" variant="caption" color="text.secondary">
                                {t('DOCUMENTS_INDEX_LABEL_REASON')}
                            </Typography>
                            <Typography component="dd" variant="body2" sx={{ m: 0 }}>
                                {t(reasonMessageKey)}
                            </Typography>
                        </>
                    ) : null}
                    {indexState.error ? (
                        <>
                            <Typography component="dt" variant="caption" color="text.secondary">
                                {t('DOCUMENTS_INDEX_LABEL_ERROR')}
                            </Typography>
                            <Typography component="dd" variant="body2" sx={{ m: 0 }}>
                                {indexState.error}
                            </Typography>
                        </>
                    ) : null}
                </Box>
            ) : null}
            <Button
                size="small"
                variant="outlined"
                disabled={reindex.isPending}
                onClick={() => reindex.mutate()}
                sx={{ alignSelf: 'flex-start' }}
            >
                {t('DOCUMENTS_REINDEX')}
            </Button>
        </Box>
    );
};

export default DocumentIndexSection;
