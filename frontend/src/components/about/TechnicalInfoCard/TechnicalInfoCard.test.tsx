import { act, render, screen } from '@testing-library/react';
import TechnicalInfoCard from './TechnicalInfoCard';

jest.mock('@/components/ui/Card/Card', () => ({ title, children }: any) => (
	<div>
		<h2>{title}</h2>
		{children}
	</div>
));

jest.mock('@/components/providers/aboutProvider/AboutContext', () => ({
	useAbout: () => ({
		commit_hash: 'abc123',
		gin_mode: 'release',
		gin_version: '1.9',
		go_version: '1.21',
		node_version: '20',
	}),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

describe('about/TechnicalInfoCard', () => {
	beforeEach(() => {
		Object.assign(navigator, {
			clipboard: {
				writeText: jest.fn().mockResolvedValue(undefined),
			},
		});
	});

	it('copies commit hash successfully', async () => {
		render(<TechnicalInfoCard />);

		await act(async () => {
			screen.getByText('COPY').click();
		});

		expect(navigator.clipboard.writeText).toHaveBeenCalledWith('abc123');
		expect(screen.getByText('COPIED')).toBeInTheDocument();
	});

	it('logs error when copy fails', async () => {
		const errorSpy = jest.spyOn(console, 'error').mockImplementation(() => {});
		(navigator.clipboard.writeText as jest.Mock).mockRejectedValueOnce(new Error('copy failed'));

		render(<TechnicalInfoCard />);

		await act(async () => {
			screen.getByText('COPY').click();
		});

		expect(errorSpy).toHaveBeenCalled();
		errorSpy.mockRestore();
	});
});
