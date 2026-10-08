import type { ReactNode } from 'react';
import PageContainer, { type PageContainerWidth } from '@/components/layout/PageContainer';
import styles from './DomainPageLayout.module.css';

interface DomainPageLayoutProps {
    header: ReactNode;
    nav: ReactNode;
    children: ReactNode;
    width?: PageContainerWidth;
}

const DomainPageLayout = ({ header, nav, children, width }: DomainPageLayoutProps) => (
    <PageContainer width={width}>
        {header}
        {nav}
        <div className={styles.contentArea}>{children}</div>
    </PageContainer>
);

export default DomainPageLayout;
