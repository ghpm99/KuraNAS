import { fireEvent, render, screen } from '@testing-library/react';
import VacuumRoomList from './VacuumRoomList';

jest.mock('@/components/i18n/provider/i18nContext', () =>
    require('./testSupport/interpolatingI18nMock').interpolatingI18nMock
);

const rooms = [
    { id: 1, name: null, area: 3.8 },
    { id: 2, name: 'Cozinha', area: 12 },
];

describe('VacuumRoomList', () => {
    it('renders with no rooms and without any provider or service', () => {
        render(
            <VacuumRoomList
                rooms={[]}
                selectedRoomIds={[]}
                onToggleRoom={jest.fn()}
                onRenameRoom={jest.fn()}
            />
        );

        expect(screen.getByText('VACUUM_ROOMS_EMPTY')).toBeInTheDocument();
    });

    it('shows the given name, the numbered fallback and the area', () => {
        render(
            <VacuumRoomList
                rooms={rooms}
                selectedRoomIds={[]}
                onToggleRoom={jest.fn()}
                onRenameRoom={jest.fn()}
            />
        );

        expect(screen.getByText('VACUUM_ROOM_DEFAULT_NAME:1')).toBeInTheDocument();
        expect(screen.getByText('Cozinha')).toBeInTheDocument();
        expect(screen.getByText('VACUUM_ROOM_AREA:3.8')).toBeInTheDocument();
    });

    it('reflects the selection in the checkboxes and toggles by keyboard-reachable checkbox', () => {
        const onToggleRoom = jest.fn();
        render(
            <VacuumRoomList
                rooms={rooms}
                selectedRoomIds={[2]}
                onToggleRoom={onToggleRoom}
                onRenameRoom={jest.fn()}
            />
        );

        const [firstCheckbox, secondCheckbox] = screen.getAllByRole('checkbox');
        expect(firstCheckbox).not.toBeChecked();
        expect(secondCheckbox).toBeChecked();

        fireEvent.click(firstCheckbox as HTMLElement);
        expect(onToggleRoom).toHaveBeenCalledWith(1);
    });

    it('asks to rename the room from its own button', () => {
        const onRenameRoom = jest.fn();
        render(
            <VacuumRoomList
                rooms={rooms}
                selectedRoomIds={[]}
                onToggleRoom={jest.fn()}
                onRenameRoom={onRenameRoom}
            />
        );

        fireEvent.click(screen.getByRole('button', { name: 'VACUUM_ROOM_RENAME_LABEL:Cozinha' }));

        expect(onRenameRoom).toHaveBeenCalledWith(rooms[1]);
    });
});
