import { useState } from 'react';
import { useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { Alert, Drawer } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import PageContainer from '@/components/layout/PageContainer';
import PageHeader from '@/components/layout/PageHeader';
import { buildDocumentsListUrl, buildDocumentViewerUrl, buildFilesUrl } from '@/app/routes';
import FileDetails from '@/features/files/fileDetails';
import useFileOperations from '@/features/files/fileActions/useFileOperations';
import DocumentIndexSection from './DocumentIndexSection';
import DocumentViewerContent from './DocumentViewerContent';
import DocumentViewerLoading from './DocumentViewerLoading';
import DocumentViewerToolbar from './DocumentViewerToolbar';
import useDocumentViewerNavigation, {
    type DocumentViewerEntry,
} from './useDocumentViewerNavigation';
import useViewerFileFallback from './useViewerFileFallback';
import { parsePageNumber } from './parsePageNumber';
import styles from './DocumentViewer.module.css';

const DocumentViewerScreen = () => {
    const { t } = useI18n();
    const navigate = useNavigate();
    const { fileId: rawFileId } = useParams();
    const [searchParams] = useSearchParams();
    const { toggleFavorites } = useFileOperations();
    const [isDetailsOpen, setIsDetailsOpen] = useState(false);

    const fileId = Number(rawFileId);
    const listSearch = searchParams.get('from') ?? '';
    const pageNumber = parsePageNumber(searchParams.get('page'));

    const { currentEntry, previousEntry, nextEntry } = useDocumentViewerNavigation(
        fileId,
        listSearch
    );
    const fallbackFile = useViewerFileFallback(fileId, currentEntry === undefined);
    const file = currentEntry?.file ?? fallbackFile.data;

    const openEntry = (entry: DocumentViewerEntry | undefined) =>
        entry
            ? () =>
                  navigate(buildDocumentViewerUrl(entry.file.id, listSearch, entry.pageNumber), {
                      replace: true,
                  })
            : undefined;

    const renderBody = () => {
        if (file) return <DocumentViewerContent file={file} pageNumber={pageNumber} />;
        if (fallbackFile.isError || !Number.isInteger(fileId)) {
            return <Alert severity="error">{t('DOCUMENTS_VIEWER_LOAD_ERROR')}</Alert>;
        }
        return <DocumentViewerLoading label={t('DOCUMENTS_VIEWER_LOADING')} />;
    };

    return (
        <PageContainer>
            <PageHeader
                title={file?.name ?? t('NAV_DOCUMENTS')}
                subtitle={file?.parent_path}
                actions={
                    <DocumentViewerToolbar
                        fileId={file?.id}
                        fileName={file?.name}
                        isStarred={Boolean(file?.starred)}
                        onBack={() => navigate(buildDocumentsListUrl(listSearch))}
                        onPrevious={openEntry(previousEntry)}
                        onNext={openEntry(nextEntry)}
                        onToggleStar={() => {
                            if (file) void toggleFavorites([file]);
                        }}
                        onOpenFolder={() => navigate(buildFilesUrl(file?.parent_path ?? ''))}
                        onOpenDetails={() => setIsDetailsOpen(true)}
                    />
                }
            />
            {renderBody()}
            <Drawer
                anchor="right"
                open={isDetailsOpen && file !== undefined}
                onClose={() => setIsDetailsOpen(false)}
            >
                {file ? (
                    <div className={styles.detailsDrawerContent}>
                        <FileDetails file={file} onClose={() => setIsDetailsOpen(false)} />
                        <DocumentIndexSection fileId={file.id} />
                    </div>
                ) : null}
            </Drawer>
        </PageContainer>
    );
};

export default DocumentViewerScreen;
