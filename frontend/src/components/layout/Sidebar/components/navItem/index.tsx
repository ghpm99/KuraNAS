import { ListItemButton, ListItemIcon, ListItemText } from '@mui/material';
import { Link, useLocation } from 'react-router-dom';

interface NavItemProps {
	href: string;
	icon: React.ReactNode;
	children: React.ReactNode;
}

const NavItem = ({ href, icon, children }: NavItemProps) => {
	const { pathname } = useLocation();

	return (
		<ListItemButton
			component={Link}
			to={href}
			selected={href === pathname}
			sx={{ borderRadius: 1, mb: 0.5 }}
		>
			<ListItemIcon sx={{ minWidth: 36 }}>{icon}</ListItemIcon>
			<ListItemText primary={children} primaryTypographyProps={{ variant: 'body2' }} />
		</ListItemButton>
	);
};

export default NavItem;
