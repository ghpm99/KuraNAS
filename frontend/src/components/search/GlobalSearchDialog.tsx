import { CircularProgress, Dialog, DialogContent, InputBase, useMediaQuery, useTheme } from '@mui/material';
import { Aperture, ArrowRightLeft, Folder, Image, Music2, Search, Video } from 'lucide-react';
import ErrorState from '@/components/errorState/errorState';
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
    isUpdating?: boolean;
    hasSearchError?: boolean;
    searchErrorMessage?: string;
    onRetry?: () => void;
    suggestion?: string;
    activeItemId: string;
    shortcut: string;
    showEmptyState: boolean;
    onClose: () => void;
    onQueryChange: (value: string) => void;
    onInputKeyDown: (event: React.KeyboardEvent<HTMLInputElement | HTMLTextAreaElement>) => void;
    onItemHover: (itemId: string) => void;
    onItemSelect: (item: SearchDialogItem) => void;
    onItemSecondaryAction?: (item: SearchDialogItem) => void;
}

const listboxId = 'global-search-listbox';

const buildOptionId = (itemId: string) => `global-search-option-${itemId}`;

const getItemIcon = (kind: SearchItemKind) => {
    switch (kind) {
        case 'folder':
            return <Folder size={18} />;
        case 'track':
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
    isUpdating = false,
    hasSearchError = false,
    searchErrorMessage = '',
    onRetry,
    suggestion = '',
    activeItemId,
    shortcut,
    showEmptyState,
    onClose,
    onQueryChange,
    onInputKeyDown,
    onItemHover,
    onItemSelect,
    onItemSecondaryAction,
}: GlobalSearchDialogProps) => {
    const { t } = useI18n();
    const theme = useTheme();
    const isPhoneViewport = useMediaQuery(theme.breakpoints.down('sm'));
    const hasOptions = sections.some((section) => section.items.length > 0);
    const activeOptionId = activeItemId ? buildOptionId(activeItemId) : undefined;

    return (
        <Dialog
            open={open}
            onClose={onClose}
            fullScreen={isPhoneViewport}
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
                        inputProps={{
                            'aria-label': t('GLOBAL_SEARCH_OPEN'),
                            role: 'combobox',
                            'aria-autocomplete': 'list',
                            'aria-expanded': hasOptions,
                            'aria-controls': listboxId,
                            'aria-activedescendant': activeOptionId,
                        }}
                    />
                    {isFetching ? (
                        <CircularProgress size={18} />
                    ) : (
                        <span className={styles.shortcut}>{shortcut}</span>
                    )}
                </div>

                <div className={styles.results} aria-busy={isUpdating}>
                    {isUpdating ? (
                        <span className={styles.updating} role="status">
                            {t('GLOBAL_SEARCH_UPDATING')}
                        </span>
                    ) : null}
                    {hasSearchError ? (
                        <ErrorState
                            title={t('GLOBAL_SEARCH_ERROR_TITLE')}
                            backendMessage={searchErrorMessage || undefined}
                            onRetry={onRetry}
                        />
                    ) : null}
                    {suggestion ? (
                        <div className={styles.suggestion} role="status">
                            <span className={styles.sectionTitle}>{t('AI_SEARCH_SUGGESTION')}</span>
                            <p>{suggestion}</p>
                        </div>
                    ) : null}
                    <div id={listboxId} role="listbox" aria-label={t('GLOBAL_SEARCH_OPEN')}>
                        {sections.map((section) => (
                            <div
                                key={section.id}
                                role="group"
                                aria-labelledby={`global-search-section-${section.id}`}
                                className={styles.section}
                            >
                                <span
                                    id={`global-search-section-${section.id}`}
                                    className={styles.sectionTitle}
                                >
                                    {section.title}
                                </span>
                                {section.items.map((item) => {
                                    const isActive = item.id === activeItemId;
                                    const itemClassName = isActive
                                        ? `${styles.item} ${styles.itemActive}`
                                        : styles.item;

                                    return (
                                        <div
                                            key={item.id}
                                            id={buildOptionId(item.id)}
                                            role="option"
                                            aria-selected={isActive}
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
                                            {item.secondaryAction ? (
                                                <button
                                                    type="button"
                                                    tabIndex={-1}
                                                    className={styles.itemSecondaryAction}
                                                    onClick={(event) => {
                                                        event.stopPropagation();
                                                        onItemSecondaryAction?.(item);
                                                    }}
                                                >
                                                    {item.secondaryAction.label}
                                                </button>
                                            ) : null}
                                        </div>
                                    );
                                })}
                            </div>
                        ))}
                    </div>

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
