import ActivityDiaryLayout from '@/components/activityDiary/activityDiaryLayout';
import ActionBar from '@/components/activityDiary/ActivityDiaryActionBar';
import ActivityDiaryForm from '@/components/activityDiary/ActivityDiaryForm';
import List from '@/components/activityDiary/ActivityList';
import Summary from '@/components/activityDiary/ActivitySummary';
import style from './activityDiary.module.css';

const ActivityDiaryPage = () => {
	return (
		<ActivityDiaryLayout>
			<div className={style['content']}>
				<ActionBar />
				<ActivityDiaryForm />
				<Summary />
				<List />
			</div>
		</ActivityDiaryLayout>
	);
};

export default ActivityDiaryPage;
