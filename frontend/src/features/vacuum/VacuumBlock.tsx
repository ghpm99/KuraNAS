import { CircularProgress } from '@mui/material';
import { useId, type ReactNode } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import ErrorState from '@/components/errorState/errorState';
import { resolveVacuumFailureMessage } from './vacuumFailures';
import styles from './VacuumBlock.module.css';

type VacuumBlockProps = {
    title: string;
    children: ReactNode;
    isLoading?: boolean;
    hasLoadError?: boolean;
    loadError?: unknown;
    onRetry?: () => void;
    className?: string;
};

const VacuumBlock = ({
    title,
    children,
    isLoading = false,
    hasLoadError = false,
    loadError,
    onRetry,
    className,
}: VacuumBlockProps) => {
    const { t } = useI18n();
    const titleId = useId();

    const renderBody = () => {
        if (isLoading) {
            return (
                <div className={styles.loading} role="status">
                    <CircularProgress size={20} />
                    <span>{t('LOADING')}</span>
                </div>
            );
        }
        if (hasLoadError) {
            return (
                <ErrorState
                    title={t('VACUUM_BLOCK_LOAD_ERROR')}
                    backendMessage={resolveVacuumFailureMessage(
                        loadError,
                        t('CONNECTION_BANNER_UNAVAILABLE')
                    )}
                    onRetry={onRetry}
                />
            );
        }
        return children;
    };

    return (
        <section
            className={className ? `${styles.block} ${className}` : styles.block}
            aria-labelledby={titleId}
        >
            <h2 id={titleId} className={styles.title}>
                {title}
            </h2>
            {renderBody()}
        </section>
    );
};

export default VacuumBlock;
