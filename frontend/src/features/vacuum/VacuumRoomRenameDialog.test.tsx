import { fireEvent, render, screen } from '@testing-library/react';
import VacuumRoomRenameDialog from './VacuumRoomRenameDialog';

jest.mock('@/components/i18n/provider/i18nContext', () =>
    require('./testSupport/interpolatingI18nMock').interpolatingI18nMock
);

const room = { id: 4, name: 'Sala', area: 10 };

describe('VacuumRoomRenameDialog', () => {
    it('renders closed without any provider or service', () => {
        render(
            <VacuumRoomRenameDialog room={null} isSaving={false} onCancel={jest.fn()} onConfirm={jest.fn()} />
        );

        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    it('starts with the current name and limits the field to 40 characters', () => {
        render(
            <VacuumRoomRenameDialog room={room} isSaving={false} onCancel={jest.fn()} onConfirm={jest.fn()} />
        );

        const field = screen.getByLabelText('VACUUM_ROOM_RENAME_FIELD');
        expect(field).toHaveValue('Sala');
        expect(field).toHaveAttribute('maxlength', '40');
        expect(screen.getByText('VACUUM_ROOM_RENAME_HELP:4')).toBeInTheDocument();
    });

    it('starts empty for a room without a name', () => {
        render(
            <VacuumRoomRenameDialog
                room={{ id: 4, name: null, area: 1 }}
                isSaving={false}
                onCancel={jest.fn()}
                onConfirm={jest.fn()}
            />
        );

        expect(screen.getByLabelText('VACUUM_ROOM_RENAME_FIELD')).toHaveValue('');
    });

    it('confirms the trimmed name and allows an empty one', () => {
        const onConfirm = jest.fn();
        render(
            <VacuumRoomRenameDialog room={room} isSaving={false} onCancel={jest.fn()} onConfirm={onConfirm} />
        );
        const field = screen.getByLabelText('VACUUM_ROOM_RENAME_FIELD');

        fireEvent.change(field, { target: { value: '  Quarto  ' } });
        fireEvent.click(screen.getByRole('button', { name: 'VACUUM_ROOM_RENAME_SAVE' }));
        expect(onConfirm).toHaveBeenLastCalledWith('Quarto');

        fireEvent.change(field, { target: { value: '' } });
        fireEvent.click(screen.getByRole('button', { name: 'VACUUM_ROOM_RENAME_SAVE' }));
        expect(onConfirm).toHaveBeenLastCalledWith('');
    });

    it('cancels', () => {
        const onCancel = jest.fn();
        render(
            <VacuumRoomRenameDialog room={room} isSaving={false} onCancel={onCancel} onConfirm={jest.fn()} />
        );

        fireEvent.click(screen.getByRole('button', { name: 'ACTION_CANCEL' }));

        expect(onCancel).toHaveBeenCalled();
    });

    it('disables both buttons while saving', () => {
        render(<VacuumRoomRenameDialog room={room} isSaving onCancel={jest.fn()} onConfirm={jest.fn()} />);

        expect(screen.getByRole('button', { name: 'VACUUM_ROOM_RENAME_SAVE' })).toBeDisabled();
        expect(screen.getByRole('button', { name: 'ACTION_CANCEL' })).toBeDisabled();
    });
});
