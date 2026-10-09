import useI18n from '@/components/i18n/provider/i18nContext';
import styles from './ErrorTechnicalDetails.module.css';

interface ErrorTechnicalDetailsProps {
    error: Error | null;
}

const ErrorTechnicalDetails = ({ error }: ErrorTechnicalDetailsProps) => {
    const { t } = useI18n();
    if (!error?.message) return null;

    return (
        <details>
            <summary>{t('ERROR_TECHNICAL_DETAILS')}</summary>
            <pre className={styles.message}>{error.message}</pre>
        </details>
    );
};

export default ErrorTechnicalDetails;
