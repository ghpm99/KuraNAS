import { useState } from 'react';
import { ChevronDown, ChevronRight, Folder } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import DocumentsFolderChildren from './DocumentsFolderChildren';
import styles from './DocumentsFolderPanel.module.css';

interface DocumentsFolderNodeProps {
    folder: FileData;
    depth: number;
    selectedFolderId?: number;
    onSelect?: (folderId: number) => void;
}

const DocumentsFolderNode = ({
    folder,
    depth,
    selectedFolderId,
    onSelect,
}: DocumentsFolderNodeProps) => {
    const { t } = useI18n();
    const [isExpanded, setIsExpanded] = useState(false);
    const isSelected = folder.id === selectedFolderId;
    const toggleLabel = t(isExpanded ? 'DOCUMENTS_FOLDERS_COLLAPSE' : 'DOCUMENTS_FOLDERS_EXPAND', {
        name: folder.name,
    });

    return (
        <li>
            <div
                className={isSelected ? `${styles.row} ${styles.rowSelected}` : styles.row}
                style={{ paddingLeft: `${depth * 16}px` }}
            >
                <button
                    type="button"
                    className={styles.toggle}
                    aria-label={toggleLabel}
                    aria-expanded={isExpanded}
                    onClick={() => setIsExpanded((wasExpanded) => !wasExpanded)}
                >
                    {isExpanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
                </button>
                <button
                    type="button"
                    className={styles.select}
                    aria-current={isSelected ? 'true' : undefined}
                    onClick={() => onSelect?.(folder.id)}
                >
                    <Folder size={14} aria-hidden="true" />
                    <span className={styles.name}>{folder.name}</span>
                </button>
            </div>
            {isExpanded ? (
                <DocumentsFolderChildren
                    parentFolderId={folder.id}
                    depth={depth + 1}
                    selectedFolderId={selectedFolderId}
                    onSelect={onSelect}
                />
            ) : null}
        </li>
    );
};

export default DocumentsFolderNode;
