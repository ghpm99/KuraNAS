import { Star } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { formatSize } from '@/shared/utils/formatSize';
import { formatDate } from '@/utils';
import type { DocumentListItem } from '@/service/documents';
import DocumentFormatIcon from './DocumentFormatIcon';
import styles from './DocumentRow.module.css';

type DocumentRowProps = {
    document: DocumentListItem;
};

const DocumentRow = ({ document }: DocumentRowProps) => {
    const { t } = useI18n();

    return (
        <li className={styles.row}>
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
            <span className={styles.star}>
                {document.starred ? (
                    <Star size={16} fill="currentColor" aria-label={t('DOCUMENTS_STARRED')} />
                ) : null}
            </span>
        </li>
    );
};

export default DocumentRow;
