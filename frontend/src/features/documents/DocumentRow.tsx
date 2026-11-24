import useI18n from '@/components/i18n/provider/i18nContext';
import { formatSize } from '@/shared/utils/formatSize';
import { formatDate } from '@/utils';
import type { DocumentListItem } from '@/service/documents';
import DocumentFormatIcon from './DocumentFormatIcon';
import DocumentMenuButton from './DocumentMenuButton';
import DocumentSelectionCheckbox from './DocumentSelectionCheckbox';
import DocumentStarButton from './DocumentStarButton';
import type { DocumentItemInteraction } from './documentItemInteraction';
import styles from './DocumentRow.module.css';

type DocumentRowProps = {
    document: DocumentListItem;
    interaction?: DocumentItemInteraction;
};

const DocumentRow = ({ document, interaction }: DocumentRowProps) => {
    const { t } = useI18n();
    const rowClassName = interaction?.isSelected ? `${styles.row} ${styles.rowSelected}` : styles.row;

    return (
        <li className={rowClassName} onContextMenu={interaction?.onContextMenu}>
            <span className={styles.select}>
                {interaction ? (
                    <DocumentSelectionCheckbox
                        documentName={document.name}
                        isSelected={interaction.isSelected}
                        onToggle={interaction.onToggleSelection}
                    />
                ) : null}
            </span>
            <span className={styles.icon}>
                <DocumentFormatIcon format={document.format} />
            </span>
            <span className={styles.name} title={document.name}>
                {document.name}
            </span>
            <span className={styles.folder} title={document.parent_path}>
                {document.parent_path}
            </span>
            <span className={styles.size}>{formatSize(document.size)}</span>
            <span className={styles.modified}>{formatDate(document.modified_at)}</span>
            <span className={styles.actions}>
                {interaction ? (
                    <>
                        <DocumentStarButton
                            documentName={document.name}
                            isStarred={document.starred}
                            onToggle={interaction.onToggleStar}
                        />
                        <DocumentMenuButton
                            documentName={document.name}
                            onOpen={interaction.onOpenMenu}
                        />
                    </>
                ) : document.starred ? (
                    <span className={styles.star} role="img" aria-label={t('DOCUMENTS_STARRED')}>
                        ★
                    </span>
                ) : null}
            </span>
        </li>
    );
};

export default DocumentRow;
