import { Search } from 'lucide-react';
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

    return (
        <div className={styles.toolbar}>
            <div className={styles.toolbarTitle}>
                <h2>{title}</h2>
                <p>{summary}</p>
            </div>
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
