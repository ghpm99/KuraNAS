import { resolveVacuumRoomName } from './resolveVacuumRoomName';

const translate = (key: string, options?: Record<string, string>) =>
    `${key}:${options?.number ?? ''}`;

describe('resolveVacuumRoomName', () => {
    it('uses the name the owner gave', () => {
        expect(resolveVacuumRoomName({ id: 3, name: '  Cozinha ', area: 4 }, translate)).toBe('Cozinha');
    });

    it.each([null, '', '   '])('falls back to the numbered name for %p', (name) => {
        expect(resolveVacuumRoomName({ id: 3, name, area: 4 }, translate)).toBe(
            'VACUUM_ROOM_DEFAULT_NAME:3'
        );
    });

    it('falls back when the name is missing from the payload', () => {
        expect(resolveVacuumRoomName({ id: 9 } as never, translate)).toBe('VACUUM_ROOM_DEFAULT_NAME:9');
    });
});
