import ActivityDiaryProvider from '@/components/hooks/ActivityDiaryProvider';
import { UIProvider } from '@/components/hooks/UI';
import { ReactNode } from 'react';
import { Layout as LayoutComponent } from './Layout';
import ActivePageListener from '@/components/activePageListener';

const Layout = ({ children }: { children: ReactNode }) => {
	return (
		<UIProvider>
			<ActivityDiaryProvider>
				<LayoutComponent>
					<ActivePageListener />
					{children}
				</LayoutComponent>
			</ActivityDiaryProvider>
		</UIProvider>
	);
};

export default Layout;
