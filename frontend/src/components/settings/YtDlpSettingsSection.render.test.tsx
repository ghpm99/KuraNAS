import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import YtDlpSettingsSection from './YtDlpSettingsSection';

describe('YtDlpSettingsSection without mocks', () => {
	it('mounts against an absent backend and shows the load error instead of crashing', async () => {
		renderWithoutBackend(<YtDlpSettingsSection />);

		expect(screen.getByRole('heading', { level: 2 })).toBeInTheDocument();
		expect(await screen.findByRole('alert')).toBeInTheDocument();
		expect(screen.queryByText('undefined')).not.toBeInTheDocument();
	});
});
