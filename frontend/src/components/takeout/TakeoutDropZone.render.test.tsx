import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import TakeoutDropZone from './TakeoutDropZone';

describe('TakeoutDropZone without mocks', () => {
	it('mounts against an absent backend with the drop zone and file picker present', () => {
		const { container } = renderWithoutBackend(<TakeoutDropZone onSelectFile={() => undefined} />);

		expect(screen.getAllByRole('button', { name: /TAKEOUT_SELECT_FILE/ }).length).toBeGreaterThan(0);
		expect(screen.getByText('TAKEOUT_DRAG_DROP')).toBeInTheDocument();
		expect(container.querySelector('input[type="file"]')).toBeInTheDocument();
		expect(screen.queryByText('undefined')).not.toBeInTheDocument();
	});
});
