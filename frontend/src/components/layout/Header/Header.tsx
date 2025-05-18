'use client';

import { Bell, Clock, Search } from 'lucide-react';

import styles from './Header.module.css';

interface HeaderProps {
	showClock?: boolean;
	currentTime?: Date;
}

export default function Header({ showClock = false, currentTime }: HeaderProps) {
	return (
		<header className={styles.header}>
			<div className={styles.searchContainer}>
				<Search className={styles.searchIcon} />
				<input type='search' placeholder='Buscar...' className={styles.searchInput} />
			</div>
			<div className={styles.actions}>
				{showClock && currentTime && (
					<div className={styles.timeDisplay}>
						<Clock className={styles.icon} />
						<span>{currentTime.toLocaleTimeString()}</span>
					</div>
				)}
				<button className={styles.iconButton}>
					<Bell className={styles.icon} />
				</button>
				<div className={styles.avatar}>
					<img src='/avatar.jpg' alt='Avatar' width={32} height={32} />
				</div>
			</div>
		</header>
	);
}
