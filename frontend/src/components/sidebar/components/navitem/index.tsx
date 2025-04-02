
interface NavItemProps {
	href: string;
	icon: React.ReactNode;
	children: React.ReactNode;
	active?: boolean;
}

const NavItem = ({ href, icon, children, active }: NavItemProps) => {
	return (
		<a href={href} className={`nav-item ${active ? 'active' : ''}`}>
			{icon}
			<span>{children}</span>
		</a>
	);
}

export default NavItem