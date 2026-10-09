import { Link } from 'react-router-dom';
import styles from './DomainNavTabs.module.css';

export interface DomainNavTabItem {
    key: string;
    href: string;
    label: string;
    isActive: boolean;
}

interface DomainNavTabsProps {
    ariaLabel: string;
    items: DomainNavTabItem[];
}

const DomainNavTabs = ({ ariaLabel, items }: DomainNavTabsProps) => (
    <nav className={styles.nav} aria-label={ariaLabel}>
        {items.map((item) => (
            <Link
                key={item.key}
                to={item.href}
                aria-current={item.isActive ? 'page' : undefined}
                className={item.isActive ? `${styles.tab} ${styles.tabActive}` : styles.tab}
            >
                {item.label}
            </Link>
        ))}
    </nav>
);

export default DomainNavTabs;
