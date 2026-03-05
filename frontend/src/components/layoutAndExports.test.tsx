import { fireEvent, render, screen } from '@testing-library/react';
import AboutLayout from './about/aboutLayout';
import ActivityDiaryLayout from './activityDiary/activityDiaryLayout';
import FilesLayout from './files/filesLayout';
import ImagesLayout from './images/imagesLayout';
import VideoLayout from './videos/videoLayout';
import MusicLayout from './music/musicLayout';
import MusicSidebar from './music/MusicSidebar';
import NavItem from './layout/Sidebar/components/navItem';
import Button from './ui/Button/Button';

import ActionBarIndex from './actionBar';
import TabsIndex from './tabs';
import FileCardIndex from './fileCard';
import FileContentIndex from './fileContent';
import FileDetailsIndex from './fileDetails';
import ImageContentIndex from './imageContent';
import MusicContentIndex from './musicContent';
import ActivityDiaryActionBarIndex from './activityDiary/ActivityDiaryActionBar';
import ActivityDiaryFormIndex from './activityDiary/ActivityDiaryForm';
import ActivityListIndex from './activityDiary/ActivityList';
import ActivitySummaryIndex from './activityDiary/ActivitySummary';

const mockUseMusic = jest.fn();

jest.mock('@/components/providers/musicProvider/musicProvider', () => ({
	useMusic: () => mockUseMusic(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

jest.mock('react-router-dom', () => ({
	Link: ({ to, children, ...props }: any) => (
		<a href={to} {...props}>
			{children}
		</a>
	),
	useLocation: () => ({ pathname: '/images' }),
}));

describe('layout wrappers and export indexes', () => {
	it('executes wrapper layouts with children', () => {
		const child = <span>child</span>;

		expect((AboutLayout as any)({ children: child })).toBeTruthy();
		expect((ActivityDiaryLayout as any)({ children: child })).toBeTruthy();
		expect((FilesLayout as any)({ children: child })).toBeTruthy();
		expect((ImagesLayout as any)({ children: child })).toBeTruthy();
		expect((VideoLayout as any)({ children: child })).toBeTruthy();
		expect((MusicLayout as any)({ children: child })).toBeTruthy();
	});

	it('renders music sidebar and updates selected view', () => {
		const setCurrentView = jest.fn();
		mockUseMusic.mockReturnValue({ currentView: 'all', setCurrentView });

		render(<MusicSidebar />);
		fireEvent.click(screen.getByText('MUSIC_ARTISTS'));
		expect(setCurrentView).toHaveBeenCalledWith('artists');
	});

	it('renders nav item link and generic button', () => {
		render(
			<NavItem href='/images' icon={<span>icon</span>}>
				Images
			</NavItem>,
		);
		expect(screen.getByRole('link', { name: /Images/ })).toBeInTheDocument();

		const onClick = jest.fn();
		render(<Button onClick={onClick}>Run</Button>);
		fireEvent.click(screen.getByRole('button', { name: 'Run' }));
		expect(onClick).toHaveBeenCalled();
	});

	it('loads index exports', () => {
		expect(ActionBarIndex).toBeDefined();
		expect(TabsIndex).toBeDefined();
		expect(FileCardIndex).toBeDefined();
		expect(FileContentIndex).toBeDefined();
		expect(FileDetailsIndex).toBeDefined();
		expect(ImageContentIndex).toBeDefined();
		expect(MusicContentIndex).toBeDefined();
		expect(ActivityDiaryActionBarIndex).toBeDefined();
		expect(ActivityDiaryFormIndex).toBeDefined();
		expect(ActivityListIndex).toBeDefined();
		expect(ActivitySummaryIndex).toBeDefined();
	});
});
