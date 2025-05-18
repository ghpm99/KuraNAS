import { useActivityDiary } from '@/components/hooks/ActivityDiaryProvider/ActivityDiaryContext';
import style from './actionBar.module.css';

const ActivityDiaryActionBar = () => {
	const { message } = useActivityDiary();
	return (
		<div className={style['action-bar']}>
			<h1 className={style['page-title']}>Diário de Atividades</h1>
			{message && <div className={`${style['message-banner']}} ${style[message.type]}`}>{message.text}</div>}
		</div>
	);
};

export default ActivityDiaryActionBar;
