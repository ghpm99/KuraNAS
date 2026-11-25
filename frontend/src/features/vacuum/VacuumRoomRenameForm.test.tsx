import { render, screen } from '@testing-library/react';
import VacuumRoomRenameForm from './VacuumRoomRenameForm';

describe('VacuumRoomRenameForm', () => {
    it('renders without any provider or service, even for a room without a name', () => {
        render(
            <VacuumRoomRenameForm
                room={{ id: 1 } as never}
                isSaving={false}
                onCancel={jest.fn()}
                onConfirm={jest.fn()}
            />
        );

        expect(screen.getByRole('button', { name: 'VACUUM_ROOM_RENAME_SAVE' })).toBeInTheDocument();
    });
});
