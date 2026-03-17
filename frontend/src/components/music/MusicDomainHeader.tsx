import useI18n from '@/components/i18n/provider/i18nContext';
import { useMusicDomainHeader } from '@/components/music/useMusicDomainHeader';
import styles from './MusicDomainHeader.module.css';

const MusicDomainHeader = () => {
    const { t } = useI18n();
    const { titleKey, descriptionKey } = useMusicDomainHeader();

    return (
        <header className={styles.header}>
            <span className={styles.eyebrow}>{t('NAV_MUSIC')}</span>
            <h1 className={styles.title}>{t(titleKey)}</h1>
            <p className={styles.description}>{t(descriptionKey)}</p>
        </header>
    );
};

export default MusicDomainHeader;
