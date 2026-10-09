import useI18n from '@/components/i18n/provider/i18nContext';
import { buildFolderCrumbs } from '../imageFolderBreadcrumb';
import styles from './ImageFolderBreadcrumb.module.css';

type ImageFolderBreadcrumbProps = {
    selectedFolder: string;
    onSelectFolder: (folderPath: string | null) => void;
};

const ImageFolderBreadcrumb = ({ selectedFolder, onSelectFolder }: ImageFolderBreadcrumbProps) => {
    const { t } = useI18n();
    const crumbs = buildFolderCrumbs(selectedFolder);

    return (
        <nav className={styles.breadcrumb} aria-label={t('IMAGES_FOLDERS_BREADCRUMB')}>
            <button type="button" className={styles.crumb} onClick={() => onSelectFolder(null)}>
                {t('IMAGES_SECTION_FOLDERS')}
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
};

export default ImageFolderBreadcrumb;
