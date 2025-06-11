import { useActivityDiary } from '@/components/hooks/ActivityDiaryProvider/ActivityDiaryContext';
import { formatDate, formatDuration } from '@/utils';
import styles from './list.module.css';
import Card from '@/components/ui/Card/Card';
import { Copy } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';

const ActivityList = () => {
	const { data, getCurrentDuration, copyActivity } = useActivityDiary();
	const { t } = useI18n();

	return (
		<Card title={t('REGISTERED_ACTIVITIES_TITLE')} className={styles['content']}>
			{data?.entries?.items.length === 0 ? (
				<p className={styles.noActivities}>{t('NO_ACTIVITIES')}</p>
			) : (
				<div className={styles.tableContainer}>
					<table className={styles.table}>
						<thead>
							<tr>
								<th>{t('NAME')}</th>
								<th>{t('DESCRIPTION')}</th>
								<th>{t('START')}</th>
								<th>{t('END')}</th>
								<th>{t('DURATION')}</th>
								<th>{t('ACTION')}</th>
							</tr>
						</thead>
						<tbody>
							{data?.entries?.items.map((activity) => (
								<tr key={activity.id} className={activity.end_time === null ? styles.activeRow : ''}>
									<td>{activity.name}</td>
									<td>{activity.description || '-'}</td>
									<td>{formatDate(activity.start_time)}</td>
									<td>{activity.end_time.HasValue ? formatDate(activity.end_time.Value) : t('IN_PROGRESS')}</td>
									<td>
										{activity.end_time.HasValue
											? formatDuration(activity.duration)
											: formatDuration(getCurrentDuration(activity.start_time))}
									</td>
									<td>
										<div onClick={() => copyActivity(activity)} className={styles.copyButton}>
											<Copy />
										</div>
									</td>
								</tr>
							))}
						</tbody>
					</table>
				</div>
			)}
		</Card>
	);
};

export default ActivityList;
