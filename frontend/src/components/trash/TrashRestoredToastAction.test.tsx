import { fireEvent, screen } from '@testing-library/react';
import TrashRestoredToastAction from './TrashRestoredToastAction';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';

describe('TrashRestoredToastAction', () => {
    it('renders without backend and asks to view the folder on click', () => {
        const onViewInFolder = jest.fn();
        renderWithoutBackend(<TrashRestoredToastAction onViewInFolder={onViewInFolder} />);

        fireEvent.click(screen.getByRole('button'));

        expect(onViewInFolder).toHaveBeenCalled();
    });
});
