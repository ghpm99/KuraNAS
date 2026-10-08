import ActivityDiaryLayout from '@/components/activityDiary/activityDiaryLayout';
import ActivityDiaryMessage from '@/components/activityDiary/ActivityDiaryMessage';
import ActivityDiaryForm from '@/components/activityDiary/ActivityDiaryForm';
import List from '@/components/activityDiary/ActivityList';
import Summary from '@/components/activityDiary/ActivitySummary';
import PageContainer from '@/components/layout/PageContainer';
import PageHeader from '@/components/layout/PageHeader';
import useI18n from '@/components/i18n/provider/i18nContext';
import style from './activityDiary.module.css';

const ActivityDiaryPage = () => {
    const { t } = useI18n();

    return (
        <ActivityDiaryLayout>
            <PageContainer>
                <PageHeader
                    title={t('ACTIVITY_DIARY_TITLE')}
                    subtitle={t('ACTIVITY_DIARY_SUBTITLE')}
                />
                <ActivityDiaryMessage />
                <div className={style['formAndSummary']}>
                    <ActivityDiaryForm />
                    <Summary />
                </div>
                <List />
            </PageContainer>
        </ActivityDiaryLayout>
    );
};

export default ActivityDiaryPage;
