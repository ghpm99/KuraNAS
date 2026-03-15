import { Button, Skeleton } from '@mui/material';
import { Link as RouterLink } from 'react-router-dom';
import type { ReactNode } from 'react';
import styles from './HomeScreen.module.css';

interface HomeSectionShellProps {
	title: string;
	description: string;
	linkLabel?: string;
	linkTo?: string;
	isLoading: boolean;
	isEmpty: boolean;
	emptyMessage: string;
	skeletonHeight?: number;
	skeletonCount?: number;
	skeletonVariant?: 'rectangular' | 'rounded';
	className?: string;
	children: ReactNode;
}

const HomeSectionShell = ({
	title,
	description,
	linkLabel,
	linkTo,
	isLoading,
	isEmpty,
	emptyMessage,
	skeletonHeight = 72,
	skeletonCount = 3,
	skeletonVariant = 'rectangular',
	className,
	children,
}: HomeSectionShellProps) => {
	return (
		<section className={`${styles.panel}${className ? ` ${className}` : ''}`}>
			<div className={styles.sectionHeader}>
				<div>
					<h2 className={styles.sectionTitle}>{title}</h2>
					<p className={styles.sectionDescription}>{description}</p>
				</div>
				{linkLabel && linkTo ? (
					<Button component={RouterLink} to={linkTo} variant='text'>
						{linkLabel}
					</Button>
				) : null}
			</div>

			{isLoading ? (
				<div className={styles.recentList}>
					{Array.from({ length: skeletonCount }).map((_, index) => (
						<Skeleton key={index} variant={skeletonVariant} height={skeletonHeight} />
					))}
				</div>
			) : isEmpty ? (
				<div className={styles.emptyState}>
					<p className={styles.emptyTitle}>{emptyMessage}</p>
				</div>
			) : (
				children
			)}
		</section>
	);
};

export default HomeSectionShell;
