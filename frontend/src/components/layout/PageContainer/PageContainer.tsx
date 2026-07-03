import type { ReactNode } from 'react';
import styles from './PageContainer.module.css';

interface PageContainerProps {
    children: ReactNode;
    className?: string;
}

const PageContainer = ({ children, className }: PageContainerProps) => (
    <div className={className ? `${styles.container} ${className}` : styles.container}>
        {children}
    </div>
);

export default PageContainer;
