import { fireEvent, render, screen } from '@testing-library/react';
import FindByDiskPathDialog from './findByDiskPathDialog';

describe('FindByDiskPathDialog without service mocks', () => {
    it('renders closed and open without providers', () => {
        const { rerender } = render(<FindByDiskPathDialog open={false} onClose={jest.fn()} />);
        expect(screen.queryByText('FILES_FIND_BY_DISK_PATH')).toBeNull();

        rerender(<FindByDiskPathDialog open onClose={jest.fn()} />);
        expect(screen.getByText('FILES_FIND_BY_DISK_PATH')).toBeInTheDocument();
        expect(
            screen.getByRole('button', { name: 'FILES_FIND_BY_DISK_PATH_SUBMIT' })
        ).toBeDisabled();
    });

    it('shows the fallback message when the backend is unreachable', async () => {
        render(<FindByDiskPathDialog open onClose={jest.fn()} />);

        fireEvent.change(screen.getByRole('textbox'), { target: { value: '/some/path' } });
        fireEvent.click(screen.getByRole('button', { name: 'FILES_FIND_BY_DISK_PATH_SUBMIT' }));

        expect(await screen.findByText('FILES_FIND_BY_DISK_PATH_FAILED')).toBeInTheDocument();
    });
});
