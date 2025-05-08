import ActionBar from '@/components/activityDiary/ActivityDiaryActionBar';
import ActivityDiaryForm from '@/components/activityDiary/ActivityDiaryForm';
import List from '@/components/activityDiary/ActivityList';
import Summary from '@/components/activityDiary/ActivitySummary';
import ActivityDiaryProvider from '@/components/providers/ActivityDiaryProvider';
import Sidebar from '@/components/sidebar';
import style from './activityDiary.module.css';

const ActivityDiaryPage = () => {
	return (
		<ActivityDiaryProvider>
			<Sidebar>
				<></>
			</Sidebar>
			<div className={style['content']}>
				<ActionBar />
				<ActivityDiaryForm />
				<Summary />
				<List />
			</div>
		</ActivityDiaryProvider>
	);
};

export default ActivityDiaryPage;
