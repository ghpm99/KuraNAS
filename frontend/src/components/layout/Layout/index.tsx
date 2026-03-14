import { ReactNode } from 'react';
import AppShell from '../AppShell';

const Layout = ({ children }: { children: ReactNode }) => {
	return <AppShell>{children}</AppShell>;
};

export default Layout;
