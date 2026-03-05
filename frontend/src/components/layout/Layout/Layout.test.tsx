import { render, screen } from '@testing-library/react';
import { Layout } from './Layout';

const mockUseUI = jest.fn();
const mockUseActivityDiary = jest.fn();
const mockUseGlobalMusic = jest.fn();
const headerSpy = jest.fn();

jest.mock('@/components/providers/uiProvider/uiContext', () => ({ useUI: () => mockUseUI() }));
jest.mock('@/components/providers/activityDiaryProvider/ActivityDiaryContext', () => ({
	useActivityDiary: () => mockUseActivityDiary(),
}));
jest.mock('@/components/providers/GlobalMusicProvider', () => ({ useGlobalMusic: () => mockUseGlobalMusic() }));

jest.mock('../Header/Header', () => ({
	__esModule: true,
	default: (props: any) => {
		headerSpy(props);
		return <div data-testid='header'>header</div>;
	},
}));

jest.mock('../Sidebar/Sidebar', () => ({
	__esModule: true,
	default: () => <div data-testid='sidebar'>sidebar</div>,
}));

describe('layout/Layout/Layout', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockUseActivityDiary.mockReturnValue({ currentTime: new Date('2026-03-04T10:00:00Z') });
	});

	it('renders with activity page clock and queue padding', () => {
		mockUseUI.mockReturnValue({ activePage: 'activity' });
		mockUseGlobalMusic.mockReturnValue({ hasQueue: true });

		render(
			<Layout>
				<div>body</div>
			</Layout>,
		);

		expect(screen.getByText('KuraNAS')).toBeInTheDocument();
		expect(screen.getByTestId('header')).toBeInTheDocument();
		expect(screen.getByTestId('sidebar')).toBeInTheDocument();
		expect(screen.getByText('body')).toBeInTheDocument();
		expect(headerSpy).toHaveBeenCalledWith(expect.objectContaining({ showClock: true }));
	});

	it('renders without clock when page is not activity', () => {
		mockUseUI.mockReturnValue({ activePage: 'files' });
		mockUseGlobalMusic.mockReturnValue({ hasQueue: false });

		render(
			<Layout>
				<div>content</div>
			</Layout>,
		);

		expect(headerSpy).toHaveBeenCalledWith(expect.objectContaining({ showClock: false }));
	});
});
