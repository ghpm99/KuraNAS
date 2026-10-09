import { Copy, Search } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { getAnalyticsImageDuplicatesRoute } from '@/app/routes';
import useI18n from '@/components/i18n/provider/i18nContext';
import styles from '../ImageContent.module.css';

type ImageToolbarProps = {
    title: string;
    summary: string;
    search: string;
    isSearchVisible: boolean;
    onSearchChange: (value: string) => void;
};

export default function ImageToolbar({
    title,
    summary,
    search,
    isSearchVisible,
    onSearchChange,
}: ImageToolbarProps) {
    const { t } = useI18n();
    const navigate = useNavigate();

    return (
        <div className={styles.toolbar}>
            <div className={styles.toolbarTitle}>
                <h2>{title}</h2>
                <p>{summary}</p>
            </div>
            <button
                type="button"
                className={styles.toolbarAction}
                onClick={() => navigate(getAnalyticsImageDuplicatesRoute())}
            >
                <Copy size={16} />
                {t('IMAGES_DUPLICATES_ACTION')}
            </button>
            {isSearchVisible && (
                <label className={styles.search}>
                    <Search size={16} />
                    <input
                        type="search"
                        value={search}
                        onChange={(event) => onSearchChange(event.target.value)}
                        placeholder={t('IMAGES_SEARCH_PLACEHOLDER')}
                        aria-label={t('IMAGES_SEARCH_PLACEHOLDER')}
                    />
                </label>
            )}
        </div>
    );
}
