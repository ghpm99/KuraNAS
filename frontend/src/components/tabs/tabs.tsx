import { useState } from 'react';

const Tabs = () => {
	const [activeTab, setActiveTab] = useState('recent');

	return (
		<div className='tabs'>
			<div className='tabs-list'>
				<button className={`tab ${activeTab === 'recent' ? 'active' : ''}`} onClick={() => setActiveTab('recent')}>
					Recentes
				</button>
				<button className={`tab ${activeTab === 'starred' ? 'active' : ''}`} onClick={() => setActiveTab('starred')}>
					Favoritos
				</button>
			</div>
		</div>
	);
};

export default Tabs;
