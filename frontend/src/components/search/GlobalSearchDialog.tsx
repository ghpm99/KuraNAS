import { CircularProgress, Dialog, DialogContent, InputBase } from '@mui/material';
import { Aperture, ArrowRightLeft, Folder, Image, Music2, Search, Video } from 'lucide-react';
import type {
    SearchDialogItem,
    SearchDialogSection,
    SearchItemKind,
} from './useGlobalSearchProvider';
import useI18n from '@/components/i18n/provider/i18nContext';
import styles from './GlobalSearchDialog.module.css';

interface GlobalSearchDialogProps {
    open: boolean;
    query: string;
    sections: SearchDialogSection[];
    isFetching: boolean;
    activeItemId: string;
    shortcut: string;
    showEmptyState: boolean;
    onClose: () => void;
    onQueryChange: (value: string) => void;
    onInputKeyDown: (event: React.KeyboardEvent<HTMLInputElement | HTMLTextAreaElement>) => void;
    onItemHover: (itemId: string) => void;
    onItemSelect: (item: SearchDialogItem) => void;
}

const getItemIcon = (kind: SearchItemKind) => {
    switch (kind) {
        case 'folder':
            return <Folder size={18} />;
        case 'artist':
        case 'album':
        case 'playlist':
            return <Music2 size={18} />;
        case 'video':
            return <Video size={18} />;
        case 'image':
            return <Image size={18} />;
        case 'action':
            return <ArrowRightLeft size={18} />;
        case 'file':
        default:
            return <Aperture size={18} />;
    }
};

const GlobalSearchDialog = ({
    open,
    query,
    sections,
    isFetching,
    activeItemId,
    shortcut,
    showEmptyState,
    onClose,
    onQueryChange,
    onInputKeyDown,
    onItemHover,
    onItemSelect,
}: GlobalSearchDialogProps) => {
    const { t } = useI18n();

    return (
        <Dialog
            open={open}
            onClose={onClose}
            fullWidth
            maxWidth="md"
            PaperProps={{ className: styles.dialogPaper }}
        >
            <DialogContent className={styles.content}>
                <div className={styles.searchField}>
                    <Search size={18} className={styles.searchIcon} />
                    <InputBase
                        autoFocus
                        value={query}
                        onChange={(event) => onQueryChange(event.target.value)}
                        onKeyDown={onInputKeyDown}
                        placeholder={t('GLOBAL_SEARCH_PLACEHOLDER')}
                        className={styles.searchInput}
                        inputProps={{ 'aria-label': t('GLOBAL_SEARCH_OPEN') }}
                    />
                    {isFetching ? (
                        <CircularProgress size={18} />
                    ) : (
                        <span className={styles.shortcut}>{shortcut}</span>
                    )}
                </div>

                <div className={styles.results}>
                    {sections.map((section) => (
                        <div key={section.id} className={styles.section}>
                            <span className={styles.sectionTitle}>{section.title}</span>
                            {section.items.map((item) => {
                                const itemClassName =
                                    item.id === activeItemId
                                        ? `${styles.item} ${styles.itemActive}`
                                        : styles.item;

                                return (
                                    <button
                                        key={item.id}
                                        type="button"
                                        className={itemClassName}
                                        onMouseEnter={() => onItemHover(item.id)}
                                        onClick={() => onItemSelect(item)}
                                    >
                                        <span className={styles.itemIcon}>
                                            {getItemIcon(item.kind)}
                                        </span>
                                        <span className={styles.itemBody}>
                                            <span className={styles.itemLabel}>{item.label}</span>
                                            <span className={styles.itemDescription}>
                                                {item.description}
                                            </span>
                                        </span>
                                        {item.meta ? (
                                            <span className={styles.itemMeta}>{item.meta}</span>
                                        ) : null}
                                    </button>
                                );
                            })}
                        </div>
                    ))}

                    {showEmptyState ? (
                        <div className={styles.emptyState}>
                            <h3>{t('GLOBAL_SEARCH_EMPTY_TITLE')}</h3>
                            <p>{t('GLOBAL_SEARCH_EMPTY_DESCRIPTION')}</p>
                        </div>
                    ) : null}
                </div>
            </DialogContent>
        </Dialog>
    );
};

export default GlobalSearchDialog;
