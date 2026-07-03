import ActivityDiaryProvider from '../providers/activityDiaryProvider';

const ActivityDiaryLayout = ({ children }: { children: React.ReactNode }) => {
    return <ActivityDiaryProvider>{children}</ActivityDiaryProvider>;
};

export default ActivityDiaryLayout;
