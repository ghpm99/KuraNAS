import { appRoutes } from '@/app/routes';
import { ListItemButton, ListItemIcon, ListItemText, Tooltip } from '@mui/material';
import { Link, useLocation } from 'react-router-dom';
import styles from './NavItem.module.css';

interface NavItemProps {
    href: string;
    icon: React.ReactNode;
    label: string;
    isIconOnly?: boolean;
    onClick?: () => void;
}

const NavItem = ({ href, icon, label, isIconOnly = false, onClick }: NavItemProps) => {
    const { pathname } = useLocation();
    const isSelected =
        href === appRoutes.home
            ? pathname === href
            : pathname === href || pathname.startsWith(`${href}/`);
    const classNames = [styles.navItem];
    if (isSelected) classNames.push(styles.selected);
    if (isIconOnly) classNames.push(styles.iconOnly);

    return (
        <Tooltip title={label} placement="right" disableHoverListener={!isIconOnly}>
            <ListItemButton
                component={Link}
                to={href}
                selected={isSelected}
                onClick={onClick}
                className={classNames.join(' ')}
                aria-label={isIconOnly ? label : undefined}
            >
                <ListItemIcon className={styles.icon}>{icon}</ListItemIcon>
                {!isIconOnly && (
                    <ListItemText primary={<span className={styles.label}>{label}</span>} />
                )}
            </ListItemButton>
        </Tooltip>
    );
};

export default NavItem;
