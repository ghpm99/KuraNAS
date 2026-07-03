import type { ReactNode } from 'react';
import PageContainer from '@/components/layout/PageContainer';
import styles from './DomainPageLayout.module.css';

interface DomainPageLayoutProps {
    header: ReactNode;
    nav: ReactNode;
    children: ReactNode;
}

const DomainPageLayout = ({ header, nav, children }: DomainPageLayoutProps) => (
    <PageContainer>
        {header}
        {nav}
        <div className={styles.contentArea}>{children}</div>
    </PageContainer>
);

export default DomainPageLayout;
