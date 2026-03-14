import { ListItemButton, ListItemIcon, ListItemText } from '@mui/material';
import { Link, useLocation } from 'react-router-dom';
import styles from './NavItem.module.css';

interface NavItemProps {
	href: string;
	icon: React.ReactNode;
	children: React.ReactNode;
	onClick?: () => void;
}

const NavItem = ({ href, icon, children, onClick }: NavItemProps) => {
	const { pathname } = useLocation();
	const className = href === pathname ? `${styles.navItem} ${styles.selected}` : styles.navItem;

	return (
		<ListItemButton
			component={Link}
			to={href}
			selected={href === pathname}
			onClick={onClick}
			className={className}
		>
			<ListItemIcon className={styles.icon}>{icon}</ListItemIcon>
			<ListItemText primary={<span className={styles.label}>{children}</span>} />
		</ListItemButton>
	);
};

export default NavItem;
