import { useState } from 'react';
import useI18n from '../i18n/provider/i18nContext';
import useFile from '../hooks/fileProvider/fileContext';
import { FileType } from '@/utils';
import './tabs.css';

const Tabs = () => {
	const [activeTab, setActiveTab] = useState('all');
	const { t } = useI18n();
	const { selectedItem } = useFile();

	if (selectedItem?.type === FileType.File) {
		return <></>;
	}
	return (
		<div className='tabs'>
			<div className='tabs-list'>
				<button className={`tab ${activeTab === 'all' ? 'active' : ''}`} onClick={() => setActiveTab('all')}>
					{t('ALL_FILES')}
				</button>
				<button className={`tab ${activeTab === 'recent' ? 'active' : ''}`} onClick={() => setActiveTab('recent')}>
					{t('RECENT_FILES')}
				</button>
				<button className={`tab ${activeTab === 'starred' ? 'active' : ''}`} onClick={() => setActiveTab('starred')}>
					{t('STARRED_FILES')}
				</button>
			</div>
		</div>
	);
};

export default Tabs;
