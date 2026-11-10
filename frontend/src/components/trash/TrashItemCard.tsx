import useI18n from '@/components/i18n/provider/i18nContext';
import { formatSize } from '@/shared/utils/formatSize';
import type { TrashItem } from '@/types/trash';
import { Button, Chip } from '@mui/material';
import { ArchiveRestore, File, Folder, Trash2 } from 'lucide-react';
import styles from './TrashScreen.module.css';
import { daysUntilExpiry, trashItemName, trashItemOriginFolder } from './trashItemPresentation';

type TrashItemCardProps = {
    item: TrashItem;
    isBusy: boolean;
    onRestore: (itemId: number) => void;
    onDeleteForever: (itemId: number) => void;
};

const expiryLabel = (
    remainingDays: number,
    t: (key: string, options?: Record<string, string>) => string
): string => {
    if (remainingDays === 0) {
        return t('TRASH_EXPIRES_NEXT_CLEANUP');
    }
    if (remainingDays === 1) {
        return t('TRASH_EXPIRES_IN_ONE_DAY');
    }
    return t('TRASH_EXPIRES_IN_DAYS', { days: String(remainingDays) });
};

const TrashItemCard = ({ item, isBusy, onRestore, onDeleteForever }: TrashItemCardProps) => {
    const { t } = useI18n();
    const remainingDays = daysUntilExpiry(item.expires_at, new Date());
    const ItemKindIcon = item.is_directory ? Folder : File;
    const itemKindLabel = t(item.is_directory ? 'TRASH_KIND_FOLDER' : 'TRASH_KIND_FILE');

    return (
        <article className={styles.item}>
            <div className={styles.itemIdentity}>
                <ItemKindIcon size={24} role="img" aria-label={itemKindLabel} />
                <div className={styles.itemInfo}>
                    <p className={styles.itemName}>{trashItemName(item)}</p>
                    <p className={styles.itemOrigin}>
                        {t('TRASH_ORIGIN_LABEL')}: {trashItemOriginFolder(item)}
                    </p>
                    <div className={styles.itemMeta}>
                        <Chip
                            size="small"
                            variant="outlined"
                            label={`${t('TRASH_DELETED_AT_LABEL')}: ${new Date(item.deleted_at).toLocaleString()}`}
                        />
                        <Chip
                            size="small"
                            variant="outlined"
                            label={`${t('TRASH_SIZE_LABEL')}: ${formatSize(item.size ?? 0)}`}
                        />
                        {remainingDays !== undefined && (
                            <Chip
                                size="small"
                                variant="outlined"
                                color={remainingDays <= 1 ? 'warning' : 'default'}
                                label={expiryLabel(remainingDays, t)}
                            />
                        )}
                    </div>
                </div>
            </div>
            <div className={styles.itemActions}>
                <Button
                    variant="contained"
                    disabled={isBusy}
                    onClick={() => onRestore(item.id)}
                    startIcon={<ArchiveRestore size={16} />}
                >
                    {t('TRASH_RESTORE_BUTTON')}
                </Button>
                <Button
                    variant="outlined"
                    color="error"
                    disabled={isBusy}
                    onClick={() => onDeleteForever(item.id)}
                    startIcon={<Trash2 size={16} />}
                >
                    {t('TRASH_DELETE_FOREVER_BUTTON')}
                </Button>
            </div>
        </article>
    );
};

export default TrashItemCard;
