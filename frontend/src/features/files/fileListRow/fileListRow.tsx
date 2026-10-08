import type { MouseEvent } from 'react';
import { Checkbox, IconButton } from '@mui/material';
import { EllipsisVertical } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import ColdTierIndicator from '@/features/files/coldTierIndicator/coldTierIndicator';
import styles from './fileListRow.module.css';

type FileListRowProps = {
    title: string;
    metadata: string;
    secondaryText?: string;
    thumbnail: string;
    onClick: (event: MouseEvent<HTMLElement>) => void;
    starred?: boolean;
    onClickStar?: () => void;
    isCold?: boolean;
    isSelected?: boolean;
    isSelectionActive?: boolean;
    onToggleSelection?: (event: MouseEvent<HTMLElement>) => void;
    onOpenMenu?: (event: MouseEvent<HTMLElement>) => void;
    onContextMenu?: (event: MouseEvent<HTMLElement>) => void;
};

const FileListRow = ({
    title,
    metadata,
    secondaryText,
    thumbnail,
    onClick,
    starred = false,
    onClickStar,
    isCold = false,
    isSelected = false,
    isSelectionActive = false,
    onToggleSelection,
    onOpenMenu,
    onContextMenu,
}: FileListRowProps) => {
    const { t } = useI18n();
    const rowClassName = isSelected ? `${styles.listRow} ${styles.listRowSelected}` : styles.listRow;
    const selectionClassName =
        isSelected || isSelectionActive
            ? `${styles.selectionControl} ${styles.selectionControlVisible}`
            : styles.selectionControl;

    return (
        <div className={rowClassName} onContextMenu={onContextMenu}>
            <div className={selectionClassName}>
                {onToggleSelection ? (
                    <Checkbox
                        size="small"
                        checked={isSelected}
                        onClick={(event) => {
                            event.stopPropagation();
                            onToggleSelection(event);
                        }}
                        slotProps={{
                            input: {
                                readOnly: true,
                                'aria-label': t('FILES_SELECT_ITEM', { name: title }),
                            },
                        }}
                        sx={{ p: 0.5 }}
                    />
                ) : null}
            </div>
            <button type="button" className={styles.listButton} onClick={onClick} aria-label={title}>
                <img src={thumbnail} alt={title} loading="lazy" className={styles.listThumbnail} />
                <div className={styles.listContent}>
                    <span className={styles.listTitle}>
                        {title}
                        {isCold ? (
                            <>
                                {' '}
                                <ColdTierIndicator />
                            </>
                        ) : null}
                    </span>
                    {secondaryText ? (
                        <span className={styles.listMetadata}>{secondaryText}</span>
                    ) : null}
                    <span className={styles.listMetadata}>{metadata}</span>
                </div>
            </button>
            <div className={styles.rowActions}>
                <button type="button" className={styles.listStarButton} onClick={onClickStar}>
                    {starred ? '★' : '☆'}
                </button>
                {onOpenMenu ? (
                    <IconButton
                        size="small"
                        aria-label={t('FILES_ITEM_MENU', { name: title })}
                        onClick={onOpenMenu}
                    >
                        <EllipsisVertical size={16} />
                    </IconButton>
                ) : null}
            </div>
        </div>
    );
};

export default FileListRow;
