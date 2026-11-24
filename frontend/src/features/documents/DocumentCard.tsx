import { Star } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { formatDate } from '@/utils';
import type { DocumentListItem } from '@/service/documents';
import DocumentFormatIcon from './DocumentFormatIcon';
import DocumentMenuButton from './DocumentMenuButton';
import DocumentSelectionCheckbox from './DocumentSelectionCheckbox';
import DocumentStarButton from './DocumentStarButton';
import type { DocumentItemInteraction } from './documentItemInteraction';
import styles from './DocumentCard.module.css';

type DocumentCardProps = {
    document: DocumentListItem;
    interaction?: DocumentItemInteraction;
};

const cardIconSize = 56;

const DocumentCard = ({ document, interaction }: DocumentCardProps) => {
    const { t } = useI18n();
    const cardClassName = interaction?.isSelected
        ? `${styles.card} ${styles.cardSelected}`
        : styles.card;

    return (
        <li className={cardClassName} onContextMenu={interaction?.onContextMenu}>
            {interaction ? (
                <span className={styles.select}>
                    <DocumentSelectionCheckbox
                        documentName={document.name}
                        isSelected={interaction.isSelected}
                        onToggle={interaction.onToggleSelection}
                    />
                </span>
            ) : null}
            <span className={styles.icon}>
                <DocumentFormatIcon format={document.format} size={cardIconSize} />
            </span>
            <span className={styles.name} title={document.name}>
                {document.name}
            </span>
            <span className={styles.folder} title={document.parent_path}>
                {document.parent_path}
            </span>
            <span className={styles.footer}>
                <span className={styles.modified}>{formatDate(document.modified_at)}</span>
                {interaction ? (
                    <span className={styles.actions}>
                        <DocumentStarButton
                            documentName={document.name}
                            isStarred={document.starred}
                            onToggle={interaction.onToggleStar}
                        />
                        <DocumentMenuButton
                            documentName={document.name}
                            onOpen={interaction.onOpenMenu}
                        />
                    </span>
                ) : document.starred ? (
                    <span className={styles.star}>
                        <Star size={16} fill="currentColor" aria-label={t('DOCUMENTS_STARRED')} />
                    </span>
                ) : null}
            </span>
        </li>
    );
};

export default DocumentCard;
