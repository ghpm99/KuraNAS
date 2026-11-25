import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import CapturesScreen from './CapturesScreen';

describe('CapturesScreen without mocks', () => {
	it('mounts against an absent backend and shows the load error instead of the empty state', async () => {
		renderWithoutBackend(<CapturesScreen />);

		expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument();
		expect(await screen.findByText('CAPTURES_ERROR')).toBeInTheDocument();
		expect(screen.queryByText('CAPTURES_EMPTY')).not.toBeInTheDocument();
		expect(screen.queryByText('undefined')).not.toBeInTheDocument();
	});
});
