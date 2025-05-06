import './App.css';

import Header from '@/components/header';
import ActivityDiaryPage from '@/pages/activityDiary';
import AnalyticsPage from '@/pages/analytics';
import FilePage from '@/pages/files';
import { BrowserRouter, Route, Routes } from 'react-router-dom';

export default function App() {
	return (
		<BrowserRouter>
			<div className='sidebar-header'>
				<h1 className='app-title'>KuraNAS</h1>
			</div>
			<Header />
			<Routes>
				<Route path='/' element={<FilePage />} />
				<Route path='/activity-diary' element={<ActivityDiaryPage />} />
				<Route path='/analytics' element={<AnalyticsPage />} />
			</Routes>
		</BrowserRouter>
	);
}
