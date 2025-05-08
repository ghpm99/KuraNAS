import ActionBar from '@/components/activityDiary/actionBar';
import ActivityDiaryForm from '@/components/activityDiary/form';
import List from '@/components/activityDiary/list';
import Summary from '@/components/activityDiary/summary';
import ActivityDiaryProvider from '@/components/providers/ActivityDiaryProvider';
import Sidebar from '@/components/sidebar';

const ActivityDiaryPage = () => {
	return (
		<ActivityDiaryProvider>
			<Sidebar>
				<></>
			</Sidebar>
			<div className='content'>
				<ActionBar />
				<ActivityDiaryForm />
				<Summary />
				<List />
			</div>
		</ActivityDiaryProvider>
	);
};

export default ActivityDiaryPage;
