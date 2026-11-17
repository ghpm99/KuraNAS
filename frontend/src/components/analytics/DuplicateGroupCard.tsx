import { useState } from 'react';
import { Link as RouterLink } from 'react-router-dom';
import { Checkbox, Chip, IconButton, Link, Tooltip } from '@mui/material';
import { ChevronDown, ChevronRight, FolderOpen, Trash2 } from 'lucide-react';
import { buildFilesUrl } from '@/app/routes';
import useI18n from '@/components/i18n/provider/i18nContext';
import { parentFolderOf } from '@/components/trash/trashItemPresentation';
import type { ItemSelection } from '@/shared/selection/useItemSelection';
import { formatDateTime } from '@/shared/utils/formatDateTime';
import { formatSize } from '@/shared/utils/formatSize';
import type { DuplicateCopy, DuplicateGroupDetail } from '@/types/analytics';
import styles from './DuplicateGroupCard.module.css';

type DuplicateGroupCardProps = {
    group: DuplicateGroupDetail;
    selection: ItemSelection<DuplicateCopy>;
    onDeleteCopy: (copy: DuplicateCopy) => void;
};

const tierLabelKeys: Record<string, string> = {
    hot: 'FILE_TIER_HOT',
    cold: 'FILE_TIER_COLD',
};

const folderUrlOf = (copy: DuplicateCopy): string => {
    const folderPath = parentFolderOf(copy.path);
    return buildFilesUrl(folderPath === '/' ? '' : folderPath);
};

const DuplicateGroupCard = ({ group, selection, onDeleteCopy }: DuplicateGroupCardProps) => {
    const { t } = useI18n();
    const [isExpanded, setIsExpanded] = useState(false);
    const ToggleIcon = isExpanded ? ChevronDown : ChevronRight;
    const copies = group.files ?? [];

    return (
        <li className={styles.group}>
            <button
                type="button"
                className={styles.groupHeader}
                aria-expanded={isExpanded}
                onClick={() => setIsExpanded((wasExpanded) => !wasExpanded)}
            >
                <ToggleIcon size={16} />
                <span className={styles.groupTitle}>
                    {t('ANALYTICS_DUPLICATES_GROUP_SUMMARY', {
                        copies: String(group.copies),
                        size: formatSize(group.size_bytes),
                        reclaimable: formatSize(group.reclaimable_size),
                    })}
                </span>
            </button>
            {isExpanded ? (
                <ul className={styles.copies}>
                    {copies.map((copy) => {
                        const tierLabelKey = tierLabelKeys[copy.tier];
                        return (
                            <li key={copy.id} className={styles.copy}>
                                <Checkbox
                                    size="small"
                                    checked={selection.isSelected(copy.id)}
                                    onChange={() => selection.toggle(copy)}
                                    inputProps={{
                                        'aria-label': t('ANALYTICS_DUPLICATES_SELECT_COPY', {
                                            path: copy.path,
                                        }),
                                    }}
                                />
                                <Link
                                    component={RouterLink}
                                    to={buildFilesUrl(copy.path)}
                                    className={styles.copyPath}
                                >
                                    {copy.path}
                                </Link>
                                <span className={styles.copyDate}>
                                    {formatDateTime(copy.updated_at)}
                                </span>
                                {tierLabelKey ? (
                                    <Chip size="small" label={t(tierLabelKey)} />
                                ) : null}
                                <Tooltip title={t('ANALYTICS_DUPLICATES_OPEN_FOLDER')}>
                                    <IconButton
                                        size="small"
                                        component={RouterLink}
                                        to={folderUrlOf(copy)}
                                        aria-label={t('ANALYTICS_DUPLICATES_OPEN_FOLDER_OF', {
                                            path: copy.path,
                                        })}
                                    >
                                        <FolderOpen size={16} />
                                    </IconButton>
                                </Tooltip>
                                <Tooltip title={t('DELETE')}>
                                    <IconButton
                                        size="small"
                                        aria-label={t('ANALYTICS_DUPLICATES_DELETE_COPY', {
                                            path: copy.path,
                                        })}
                                        onClick={() => onDeleteCopy(copy)}
                                    >
                                        <Trash2 size={16} />
                                    </IconButton>
                                </Tooltip>
                            </li>
                        );
                    })}
                </ul>
            ) : null}
        </li>
    );
};

export default DuplicateGroupCard;
