import { Alert } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import DocumentExtractedText from './DocumentExtractedText';
import DocumentFormattedText from './DocumentFormattedText';
import DocumentPdfFrame from './DocumentPdfFrame';
import DocumentViewerLoading from './DocumentViewerLoading';
import { resolveDocumentIndexNotice } from './documentIndexNotice';
import { resolveDocumentViewerKind } from './documentViewerKind';
import useDocumentIndexState from './useDocumentIndexState';
import styles from './DocumentViewer.module.css';

type DocumentViewerContentProps = {
    file: FileData;
    pageNumber?: number;
};

const DocumentViewerContent = ({ file, pageNumber }: DocumentViewerContentProps) => {
    const { t } = useI18n();
    const kind = resolveDocumentViewerKind(file.format);
    const indexQuery = useDocumentIndexState(file.id);
    const indexNotice = resolveDocumentIndexNotice(indexQuery.data);

    if (kind === 'pdf') {
        if (indexQuery.isPending) {
            return <DocumentViewerLoading label={t('DOCUMENTS_VIEWER_LOADING')} />;
        }
        if (indexNotice === 'protected') {
            return <Alert severity="warning">{t('DOCUMENTS_VIEWER_PROTECTED')}</Alert>;
        }
        return (
            <div className={styles.content}>
                {indexNotice === 'scanned' ? (
                    <Alert severity="info">{t('DOCUMENTS_VIEWER_SCANNED_PDF')}</Alert>
                ) : null}
                <DocumentPdfFrame fileId={file.id} title={file.name} pageNumber={pageNumber} />
            </div>
        );
    }

    if (kind === 'extractedText') {
        if (indexQuery.isPending) {
            return <DocumentViewerLoading label={t('DOCUMENTS_VIEWER_LOADING')} />;
        }
        return <DocumentExtractedText fileId={file.id} indexNotice={indexNotice} />;
    }

    return <DocumentFormattedText fileId={file.id} fileSize={file.size} kind={kind} />;
};

export default DocumentViewerContent;
