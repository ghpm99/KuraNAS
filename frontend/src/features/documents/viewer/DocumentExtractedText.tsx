import { Alert } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { DocumentIndexNotice } from './documentIndexNotice';
import DocumentViewerLoading from './DocumentViewerLoading';
import useDocumentText from './useDocumentText';
import styles from './DocumentViewer.module.css';

type DocumentExtractedTextProps = {
    fileId: number;
    indexNotice: DocumentIndexNotice | null;
};

const noticeMessageKeys: Record<DocumentIndexNotice, string> = {
    protected: 'DOCUMENTS_VIEWER_PROTECTED',
    scanned: 'DOCUMENTS_VIEWER_SCANNED',
};

const DocumentExtractedText = ({ fileId, indexNotice }: DocumentExtractedTextProps) => {
    const { t } = useI18n();
    const {
        data: documentText,
        isPending,
        isError,
    } = useDocumentText(fileId, indexNotice === null);

    if (indexNotice !== null) {
        return <Alert severity="info">{t(noticeMessageKeys[indexNotice])}</Alert>;
    }
    if (isPending) return <DocumentViewerLoading label={t('TEXT_VIEWER_LOADING')} />;
    if (isError) return <Alert severity="warning">{t('DOCUMENTS_VIEWER_TEXT_ERROR')}</Alert>;

    const extractedText = documentText?.text ?? '';
    if (extractedText === '') return <Alert severity="info">{t('DOCUMENTS_VIEWER_NO_TEXT')}</Alert>;

    return (
        <div className={styles.content}>
            <Alert severity="info">{t('DOCUMENTS_VIEWER_SIMPLIFIED')}</Alert>
            {documentText?.is_partial ? (
                <Alert severity="warning">{t('DOCUMENTS_VIEWER_PARTIAL')}</Alert>
            ) : null}
            <pre className={styles.extractedText} tabIndex={0}>
                {extractedText}
            </pre>
        </div>
    );
};

export default DocumentExtractedText;
