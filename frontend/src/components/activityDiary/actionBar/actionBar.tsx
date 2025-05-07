import { messageType } from '@/components/providers/ActivityDiaryProvider/ActivityDiaryContext';

const ActionBar = (message: { text: string; type: messageType }) => {
	return (
		<>
			{message && <div className={`message-banner ${message.type}`}>{message.text}</div>}

			<div className='action-bar'>
				<h1 className='page-title'>Diário de Atividades</h1>
			</div>
		</>
	);
};

export default ActionBar;
