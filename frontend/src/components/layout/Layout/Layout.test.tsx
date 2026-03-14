import { render, screen } from '@testing-library/react';
import { Layout } from './Layout';

const mockUseAppShell = jest.fn();
const headerSpy = jest.fn();

jest.mock('@/components/layout/AppShell/useAppShell', () => ({ useAppShell: () => mockUseAppShell() }));

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
	});

	it('renders with activity page clock and queue padding', () => {
		mockUseAppShell.mockReturnValue({ showClock: true, hasQueue: true });

		render(
			<Layout>
				<div>body</div>
			</Layout>,
		);

		expect(screen.getByTestId('header')).toBeInTheDocument();
		expect(screen.getByTestId('sidebar')).toBeInTheDocument();
		expect(screen.getByText('body')).toBeInTheDocument();
		expect(headerSpy).toHaveBeenCalledWith(expect.objectContaining({ showClock: true }));
	});

	it('renders without clock when page is not activity', () => {
		mockUseAppShell.mockReturnValue({ showClock: false, hasQueue: false });

		render(
			<Layout>
				<div>content</div>
			</Layout>,
		);

		expect(headerSpy).toHaveBeenCalledWith(expect.objectContaining({ showClock: false }));
	});
});
