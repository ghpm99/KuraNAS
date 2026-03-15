import type { ReactNode } from 'react';
import styles from './DomainPageLayout.module.css';

interface DomainPageLayoutProps {
	header: ReactNode;
	sidebar: ReactNode;
	children: ReactNode;
}

const DomainPageLayout = ({ header, sidebar, children }: DomainPageLayoutProps) => (
	<div className={styles.container}>
		<div className={styles.page}>
			{header}
			<div className={styles.domainContent}>
				{sidebar}
				<div className={styles.contentArea}>{children}</div>
			</div>
		</div>
	</div>
);

export default DomainPageLayout;
