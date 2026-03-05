import { render, screen } from '@testing-library/react';
import SystemInfoCard from './SystemInfoCard';
const mockUseAbout = jest.fn();

jest.mock('@/components/ui/Card/Card', () => ({ title, children }: any) => (
	<div>
		<h2>{title}</h2>
		{children}
	</div>
));

jest.mock('@/components/providers/aboutProvider/AboutContext', () => ({
	useAbout: () => mockUseAbout(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

describe('about/SystemInfoCard', () => {
	beforeEach(() => {
		mockUseAbout.mockReturnValue({ version: '1.2.3', platform: 'linux', lang: 'pt-BR' });
	});

	it('renders translated labels and values', () => {
		render(<SystemInfoCard />);

		expect(screen.getByText('SYSTEM_INFO_TITLE')).toBeInTheDocument();
		expect(screen.getByText('1.2.3')).toBeInTheDocument();
		expect(screen.getByText(/linux/)).toBeInTheDocument();
		expect(screen.getByText('pt-BR')).toBeInTheDocument();
	});

	it('renders windows platform icon branch', () => {
		mockUseAbout.mockReturnValue({ version: '2.0.0', platform: 'windows', lang: 'en-US' });
		render(<SystemInfoCard />);
		expect(screen.getByText(/windows/)).toBeInTheDocument();
	});
});
