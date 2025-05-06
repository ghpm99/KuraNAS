import { Link, useLocation } from 'react-router-dom';
import './navItem.css';
interface NavItemProps {
	href: string;
	icon: React.ReactNode;
	children: React.ReactNode;
}

const NavItem = ({ href, icon, children }: NavItemProps) => {
	const { pathname } = useLocation();
	return (
		<Link to={href} className={`nav-item ${href === pathname ? 'active' : ''}`}>
			{icon}
			<span>{children}</span>
		</Link>
	);
};

export default NavItem;
