import AppShell from '../layout/AppShell';
import ActivityDiaryProvider from '../providers/activityDiaryProvider';

const ActivityDiaryLayout = ({ children }: { children: React.ReactNode }) => {
	return (
		<ActivityDiaryProvider>
			<AppShell>{children}</AppShell>
		</ActivityDiaryProvider>
	);
};

export default ActivityDiaryLayout;
