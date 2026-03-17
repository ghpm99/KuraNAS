import { Link } from 'react-router-dom';
import { Button, Chip } from '@mui/material';
import { Copy, ExternalLink } from 'lucide-react';
import styles from './AboutScreen.module.css';
import { useAboutScreen } from './useAboutScreen';

const AboutScreen = () => {
    const {
        t,
        version,
        workersEnabled,
        runtimeDetails,
        buildDetails,
        tools,
        copied,
        copyCommitHash,
    } = useAboutScreen();

    return (
        <div className={styles.page}>
            <header className={styles.hero}>
                <div>
                    <p className={styles.eyebrow}>{t('ABOUT_PAGE_EYEBROW')}</p>
                    <h1 className={styles.title}>{t('ABOUT_PAGE_TITLE')}</h1>
                    <p className={styles.description}>{t('ABOUT_PAGE_DESCRIPTION')}</p>
                </div>

                <div className={styles.heroMeta}>
                    <Chip
                        label={`${t('ABOUT_RUNTIME_VERSION')}: ${version}`}
                        color="primary"
                        variant="outlined"
                    />
                    <Chip
                        label={workersEnabled ? t('ABOUT_WORKERS_ON') : t('ABOUT_WORKERS_OFF')}
                        color={workersEnabled ? 'success' : 'default'}
                        variant="outlined"
                    />
                </div>
            </header>

            <div className={styles.grid}>
                <section className={styles.panel}>
                    <div className={styles.sectionHeader}>
                        <h2 className={styles.sectionTitle}>{t('ABOUT_SECTION_RUNTIME_TITLE')}</h2>
                        <p className={styles.sectionDescription}>
                            {t('ABOUT_SECTION_RUNTIME_DESCRIPTION')}
                        </p>
                    </div>
                    <dl className={styles.detailList}>
                        {runtimeDetails.map((item) => (
                            <div key={item.label} className={styles.detailRow}>
                                <dt className={styles.detailLabel}>{item.label}</dt>
                                <dd className={styles.detailValue}>{item.value}</dd>
                            </div>
                        ))}
                    </dl>
                </section>

                <section className={styles.panel}>
                    <div className={styles.sectionHeader}>
                        <h2 className={styles.sectionTitle}>{t('ABOUT_SECTION_BUILD_TITLE')}</h2>
                        <p className={styles.sectionDescription}>
                            {t('ABOUT_SECTION_BUILD_DESCRIPTION')}
                        </p>
                    </div>
                    <dl className={styles.detailList}>
                        {buildDetails.map((item, index) => (
                            <div key={item.label} className={styles.detailRow}>
                                <dt className={styles.detailLabel}>{item.label}</dt>
                                <dd
                                    className={
                                        index === 0
                                            ? `${styles.detailValue} ${styles.commitValue}`
                                            : styles.detailValue
                                    }
                                >
                                    {index === 0 ? (
                                        <div className={styles.commitRow}>
                                            <span>{item.value}</span>
                                            <Button
                                                variant="text"
                                                size="small"
                                                onClick={() => void copyCommitHash()}
                                                startIcon={<Copy size={14} />}
                                            >
                                                {copied
                                                    ? t('ABOUT_COMMIT_COPIED')
                                                    : t('ABOUT_COPY_COMMIT')}
                                            </Button>
                                        </div>
                                    ) : (
                                        item.value
                                    )}
                                </dd>
                            </div>
                        ))}
                    </dl>
                </section>

                <section className={`${styles.panel} ${styles.toolsPanel}`}>
                    <div className={styles.sectionHeader}>
                        <h2 className={styles.sectionTitle}>{t('ABOUT_SECTION_TOOLS_TITLE')}</h2>
                        <p className={styles.sectionDescription}>
                            {t('ABOUT_SECTION_TOOLS_DESCRIPTION')}
                        </p>
                    </div>
                    <div className={styles.toolsGrid}>
                        {tools.map((tool) => (
                            <article key={tool.href} className={styles.toolCard}>
                                <div>
                                    <h3 className={styles.toolTitle}>{tool.title}</h3>
                                    <p className={styles.toolDescription}>{tool.description}</p>
                                </div>
                                <Button
                                    component={Link}
                                    to={tool.href}
                                    variant="outlined"
                                    endIcon={<ExternalLink size={14} />}
                                >
                                    {t('ABOUT_OPEN_TOOL')}
                                </Button>
                            </article>
                        ))}
                    </div>
                </section>
            </div>
        </div>
    );
};

export default AboutScreen;
