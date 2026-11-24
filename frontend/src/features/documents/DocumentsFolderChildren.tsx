import { Button, CircularProgress } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import DocumentsFolderNode from './DocumentsFolderNode';
import useDocumentsFolderChildren from './useDocumentsFolderChildren';
import styles from './DocumentsFolderPanel.module.css';

interface DocumentsFolderChildrenProps {
    parentFolderId: number;
    depth: number;
    selectedFolderId?: number;
    onSelect?: (folderId: number) => void;
}

const DocumentsFolderChildren = ({
    parentFolderId,
    depth,
    selectedFolderId,
    onSelect,
}: DocumentsFolderChildrenProps) => {
    const { t } = useI18n();
    const {
        subfolders,
        isLoading,
        isError,
        hasNextPage,
        isFetchingNextPage,
        fetchNextPage,
        retry,
    } = useDocumentsFolderChildren(parentFolderId);
    const indentation = { paddingLeft: `${depth * 16 + 8}px` };

    if (isLoading) {
        return (
            <div className={styles.feedback} style={indentation} role="status">
                <CircularProgress size={14} aria-label={t('DOCUMENTS_FOLDERS_LOADING')} />
            </div>
        );
    }

    if (isError) {
        return (
            <div className={styles.feedback} style={indentation}>
                <span>{t('DOCUMENTS_FOLDERS_LOAD_ERROR')}</span>
                <Button size="small" onClick={retry}>
                    {t('DOCUMENTS_FOLDERS_RETRY')}
                </Button>
            </div>
        );
    }

    return (
        <ul className={styles.nodeList}>
            {subfolders.map((folder) => (
                <DocumentsFolderNode
                    key={folder.id}
                    folder={folder}
                    depth={depth}
                    selectedFolderId={selectedFolderId}
                    onSelect={onSelect}
                />
            ))}
            {subfolders.length === 0 && !hasNextPage ? (
                <li className={styles.feedback} style={indentation}>
                    {t('DOCUMENTS_FOLDERS_NO_SUBFOLDERS')}
                </li>
            ) : null}
            {hasNextPage ? (
                <li style={indentation}>
                    <Button size="small" onClick={fetchNextPage} disabled={isFetchingNextPage}>
                        {t('DOCUMENTS_FOLDERS_LOAD_MORE')}
                    </Button>
                </li>
            ) : null}
        </ul>
    );
};

export default DocumentsFolderChildren;
