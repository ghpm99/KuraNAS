import useI18n from '@/components/i18n/provider/i18nContext';
import { buildVideoFolderCrumbs } from './videoFolderCrumbs';
import styles from './VideoFolderBrowser.module.css';

type VideoFolderBreadcrumbProps = {
    selectedFolder: string;
    onSelectFolder: (folderPath: string | null) => void;
};

export default function VideoFolderBreadcrumb({
    selectedFolder,
    onSelectFolder,
}: VideoFolderBreadcrumbProps) {
    const { t } = useI18n();
    const crumbs = buildVideoFolderCrumbs(selectedFolder);

    return (
        <nav className={styles.breadcrumb} aria-label={t('VIDEO_FOLDERS_BREADCRUMB')}>
            <button type="button" className={styles.crumb} onClick={() => onSelectFolder(null)}>
                {t('VIDEO_SECTION_FOLDERS')}
            </button>
            {crumbs.map((crumb, position) => {
                const isCurrent = position === crumbs.length - 1;
                return (
                    <span key={crumb.path} className={styles.step}>
                        <span className={styles.separator} aria-hidden="true">
                            /
                        </span>
                        <button
                            type="button"
                            className={
                                isCurrent ? `${styles.crumb} ${styles.current}` : styles.crumb
                            }
                            aria-current={isCurrent ? 'page' : undefined}
                            onClick={() => onSelectFolder(crumb.path)}
                        >
                            {crumb.name}
                        </button>
                    </span>
                );
            })}
        </nav>
    );
}
