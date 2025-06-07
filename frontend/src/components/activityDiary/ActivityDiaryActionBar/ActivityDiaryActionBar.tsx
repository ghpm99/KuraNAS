import { useActivityDiary } from '@/components/hooks/ActivityDiaryProvider/ActivityDiaryContext';
import style from './actionBar.module.css';
import useI18n from '@/components/i18n/provider/i18nContext';

const ActivityDiaryActionBar = () => {
	const { message } = useActivityDiary();
	const { t } = useI18n();
	return (
		<div className={style['action-bar']}>
			<h1 className={style['page-title']}>{t('ACTIVITY_DIARY_TITLE')}</h1>
			{message && <div className={`${style['message-banner']}} ${style[message.type]}`}>{message.text}</div>}
		</div>
	);
};

export default ActivityDiaryActionBar;
