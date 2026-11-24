import { Button, CircularProgress } from '@mui/material';
import { Link } from 'react-router-dom';
import { Folder } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { appRoutes } from '@/app/routes';
import DocumentsFolderChildren from './DocumentsFolderChildren';
import useDocumentsLibraryRoot from './useDocumentsLibraryRoot';
import styles from './DocumentsFolderPanel.module.css';

interface DocumentsFolderPanelProps {
    selectedFolderId?: number;
    onSelect?: (folderId: number | undefined) => void;
}

const DocumentsFolderPanel = ({ selectedFolderId, onSelect }: DocumentsFolderPanelProps) => {
    const { t } = useI18n();
    const { status, root, retry } = useDocumentsLibraryRoot();
    const isAllSelected = selectedFolderId === undefined;

    const renderTree = () => {
        if (status === 'loading') {
            return (
                <div className={styles.feedback} role="status">
                    <CircularProgress size={14} aria-label={t('DOCUMENTS_FOLDERS_LOADING')} />
                </div>
            );
        }
        if (status === 'error') {
            return (
                <div className={styles.feedback}>
                    <span>{t('DOCUMENTS_FOLDERS_LOAD_ERROR')}</span>
                    <Button size="small" onClick={retry}>
                        {t('DOCUMENTS_FOLDERS_RETRY')}
                    </Button>
                </div>
            );
        }
        if (status === 'not-configured' || !root) {
            return (
                <div className={styles.feedback}>
                    <p>{t('DOCUMENTS_FOLDERS_NOT_CONFIGURED')}</p>
                    <Button size="small" component={Link} to={appRoutes.settings}>
                        {t('DOCUMENTS_FOLDERS_OPEN_SETTINGS')}
                    </Button>
                </div>
            );
        }
        return (
            <DocumentsFolderChildren
                parentFolderId={root.id}
                depth={0}
                selectedFolderId={selectedFolderId}
                onSelect={onSelect}
            />
        );
    };

    return (
        <nav className={styles.panel} aria-label={t('DOCUMENTS_FOLDERS_TITLE')}>
            <p className={styles.title}>{t('DOCUMENTS_FOLDERS_TITLE')}</p>
            <div className={isAllSelected ? `${styles.row} ${styles.rowSelected}` : styles.row}>
                <button
                    type="button"
                    className={styles.select}
                    aria-current={isAllSelected ? 'true' : undefined}
                    onClick={() => onSelect?.(undefined)}
                >
                    <Folder size={14} aria-hidden="true" />
                    <span className={styles.name}>{t('DOCUMENTS_FOLDERS_ALL')}</span>
                </button>
            </div>
            {renderTree()}
        </nav>
    );
};

export default DocumentsFolderPanel;
