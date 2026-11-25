import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import TakeoutImportScreen from './TakeoutImportScreen';

describe('TakeoutImportScreen without mocks', () => {
	it('mounts against an absent backend with the drop zone and start control present', () => {
		const { container } = renderWithoutBackend(<TakeoutImportScreen />);

		expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument();
		expect(screen.getByText('TAKEOUT_DRAG_DROP')).toBeInTheDocument();
		expect(container.querySelector('input[type="file"]')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'TAKEOUT_START_UPLOAD' })).toBeDisabled();
		expect(screen.queryByText('undefined')).not.toBeInTheDocument();
		expect(container.textContent).not.toContain('undefined');
	});
});
