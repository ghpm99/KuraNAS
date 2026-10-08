import type { ReactNode } from 'react';
import styles from './PageContainer.module.css';

export type PageContainerWidth = 'contained' | 'full';

interface PageContainerProps {
    children: ReactNode;
    className?: string;
    width?: PageContainerWidth;
}

const PageContainer = ({ children, className, width = 'contained' }: PageContainerProps) => {
    const classNames = [styles.container];
    if (width === 'full') classNames.push(styles.containerFull);
    if (className) classNames.push(className);

    return <div className={classNames.join(' ')}>{children}</div>;
};

export default PageContainer;
