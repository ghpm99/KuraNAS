import { Star } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { formatDate } from '@/utils';
import type { DocumentListItem } from '@/service/documents';
import DocumentFormatIcon from './DocumentFormatIcon';
import styles from './DocumentCard.module.css';

type DocumentCardProps = {
    document: DocumentListItem;
};

const cardIconSize = 56;

const DocumentCard = ({ document }: DocumentCardProps) => {
    const { t } = useI18n();

    return (
        <li className={styles.card}>
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
                {document.starred ? (
                    <span className={styles.star}>
                        <Star size={16} fill="currentColor" aria-label={t('DOCUMENTS_STARRED')} />
                    </span>
                ) : null}
            </span>
        </li>
    );
};

export default DocumentCard;
