import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import ImageViewerDetailItem from './ImageViewerDetailItem';

describe('ImageViewerDetailItem (no-mock render)', () => {
    it('renders a plain label and value without link or copy button', () => {
        renderWithoutBackend(<ImageViewerDetailItem item={{ label: 'Camera', value: 'N/A' }} />);

        expect(screen.getByText('Camera')).toBeInTheDocument();
        expect(screen.getByText('N/A')).toBeInTheDocument();
        expect(screen.queryByRole('link')).not.toBeInTheDocument();
        expect(screen.queryByRole('button')).not.toBeInTheDocument();
    });
});
