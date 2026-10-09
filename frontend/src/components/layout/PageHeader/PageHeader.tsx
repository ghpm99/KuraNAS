import type { ReactNode } from 'react';
import { useDocumentTitle } from '@/components/layout/documentMetadata/useDocumentTitle';
import styles from './PageHeader.module.css';

interface PageHeaderProps {
    title: string;
    subtitle?: string;
    actions?: ReactNode;
}

const PageHeader = ({ title, subtitle, actions }: PageHeaderProps) => {
    useDocumentTitle(title);

    return (
        <header className={styles.header}>
            <div className={styles.copy}>
                <h1 className={styles.title} tabIndex={-1}>
                    {title}
                </h1>
                {subtitle ? <p className={styles.subtitle}>{subtitle}</p> : null}
            </div>
            {actions ? <div className={styles.actions}>{actions}</div> : null}
        </header>
    );
};

export default PageHeader;
