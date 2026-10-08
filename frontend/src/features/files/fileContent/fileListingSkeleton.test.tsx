import { render, screen } from '@testing-library/react';
import FileListingSkeleton from './fileListingSkeleton';

describe('FileListingSkeleton', () => {
    it('renders a busy status with the loading label and no providers', () => {
        render(<FileListingSkeleton viewMode="grid" />);

        const status = screen.getByRole('status');
        expect(status).toHaveAttribute('aria-busy', 'true');
        expect(status).toHaveTextContent('LOADING');
    });

    it('renders one placeholder per requested slot in grid mode', () => {
        const { container } = render(<FileListingSkeleton viewMode="grid" placeholderCount={5} />);

        expect(container.querySelectorAll('.MuiSkeleton-rounded')).toHaveLength(5);
    });

    it('renders row placeholders with a thumbnail square in list mode', () => {
        const { container } = render(<FileListingSkeleton viewMode="list" placeholderCount={3} />);

        expect(container.querySelectorAll('.MuiSkeleton-rounded')).toHaveLength(3);
        expect(container.querySelectorAll('.MuiSkeleton-text')).toHaveLength(6);
    });
});
