import { useState } from 'react'

const Tabs = () => {
    const [activeTab, setActiveTab] = useState('recent');

    return (
        <div className='tabs'>
						<div className='tabs-list'>
							<button
								className={`tab ${activeTab === 'recent' ? 'active' : ''}`}
								onClick={() => setActiveTab('recent')}
							>
								Recent
							</button>
							<button
								className={`tab ${activeTab === 'starred' ? 'active' : ''}`}
								onClick={() => setActiveTab('starred')}
							>
								Starred
							</button>
							<button
								className={`tab ${activeTab === 'shared' ? 'active' : ''}`}
								onClick={() => setActiveTab('shared')}
							>
								Shared
							</button>
						</div>
					</div>
    )
}

export default Tabs