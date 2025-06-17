import { FileType } from '@/utils';
import useFile from '../hooks/fileProvider/fileContext';
import useI18n from '../i18n/provider/i18nContext';
import './tabs.css';

const Tabs = () => {
	const { t } = useI18n();
	const { selectedItem, fileListFilter, setFileListFilter } = useFile();

	if (selectedItem?.type === FileType.File) {
		return <></>;
	}
	return (
		<div className='tabs'>
			<div className='tabs-list'>
				<button className={`tab ${fileListFilter === 'all' ? 'active' : ''}`} onClick={() => setFileListFilter('all')}>
					{t('ALL_FILES')}
				</button>
				<button
					className={`tab ${fileListFilter === 'recent' ? 'active' : ''}`}
					onClick={() => setFileListFilter('recent')}
				>
					{t('RECENT_FILES')}
				</button>
				<button
					className={`tab ${fileListFilter === 'starred' ? 'active' : ''}`}
					onClick={() => setFileListFilter('starred')}
				>
					{t('STARRED_FILES')}
				</button>
			</div>
		</div>
	);
};

export default Tabs;
