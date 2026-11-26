import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import CapturePromotedActions from './CapturePromotedActions';

describe('CapturePromotedActions without mocks', () => {
    it('mounts against an absent backend and shows the folder error after a failed lookup', async () => {
        renderWithoutBackend(<CapturePromotedActions captureId={7} fileId={42} />);

        expect(screen.getByRole('link', { name: 'CAPTURES_WATCH' })).toHaveAttribute(
            'href',
            '/video/42'
        );
        fireEvent.click(screen.getByRole('button', { name: 'CAPTURES_OPEN_FOLDER' }));

        expect(await screen.findByText('CAPTURES_OPEN_FOLDER_ERROR')).toBeInTheDocument();
    });
});
