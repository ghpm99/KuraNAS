import { screen } from '@testing-library/react';
import TrashPage from '@/pages/trash';
import {
    expectRendersWithoutBackend,
    renderWithoutBackend,
} from '@/shared/test/renderWithoutBackend';
import TrashConfirmDialog from './TrashConfirmDialog';
import TrashScreen from './TrashScreen';

describe('TrashScreen (no-mock render)', () => {
    it('renders without throwing when the backend is unavailable', () => {
        expectRendersWithoutBackend(<TrashScreen />);
    });

    it('settles on the load error instead of a blank screen', async () => {
        renderWithoutBackend(<TrashScreen />);

        expect(await screen.findByText('TRASH_LOAD_ERROR')).toBeInTheDocument();
    });

    it('renders the page shell without backend', () => {
        expectRendersWithoutBackend(<TrashPage />);
    });
});

describe('TrashConfirmDialog (no-mock render)', () => {
    it('renders open without backend', () => {
        expectRendersWithoutBackend(
            <TrashConfirmDialog
                isOpen
                title="title"
                message="message"
                onCancel={jest.fn()}
                onConfirm={jest.fn()}
            />
        );
    });
});
