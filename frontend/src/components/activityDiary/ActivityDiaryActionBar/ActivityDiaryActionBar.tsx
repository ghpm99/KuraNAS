import { useActivityDiary } from '@/components/providers/ActivityDiaryProvider/ActivityDiaryContext';
import style from './actionBar.module.css';

const ActivityDiaryActionBar = () => {
	const { message } = useActivityDiary();
	return (
		<div className={style['action-bar']}>
			{message && <div className={`${style['message-banner']}} ${style[message.type]}`}>{message.text}</div>}

			<h1 className={style['page-title']}>Diário de Atividades</h1>
		</div>
	);
};

export default ActivityDiaryActionBar;
