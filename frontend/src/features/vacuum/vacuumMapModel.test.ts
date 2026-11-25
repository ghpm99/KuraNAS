import { buildVacuumMapPayload } from './testSupport/buildVacuumMapPayload';
import { toVacuumMapModel } from './vacuumMapModel';

describe('toVacuumMapModel', () => {
    it('returns null for an absent or empty payload', () => {
        expect(toVacuumMapModel(undefined)).toBeNull();
        expect(toVacuumMapModel({})).toBeNull();
    });

    it('returns null when the cells are missing or do not fit the grid', () => {
        expect(toVacuumMapModel(buildVacuumMapPayload({ cells: undefined }))).toBeNull();
        expect(toVacuumMapModel(buildVacuumMapPayload({ cells: btoa('abc') }))).toBeNull();
    });

    it('returns null for a non positive resolution', () => {
        expect(toVacuumMapModel(buildVacuumMapPayload({ resolution: 0 }))).toBeNull();
    });

    it('maps a complete payload', () => {
        const model = toVacuumMapModel(buildVacuumMapPayload());

        expect(model).toMatchObject({
            mapId: 77,
            width: 4,
            height: 3,
            resolution: 50,
            originX: -100,
            originY: -50,
            dock: { x: 25, y: 25 },
            forbiddenZones: [
                [
                    [0, 0],
                    [50, 0],
                    [50, 50],
                ],
            ],
        });
        expect(model?.cells).toHaveLength(12);
    });

    it('tolerates a missing dock, missing zones and malformed vertices', () => {
        const model = toVacuumMapModel(
            buildVacuumMapPayload({
                dock: null,
                forbidden_zones: [
                    { vertices: undefined as unknown as [number, number][] },
                    { vertices: [[1, 2], ['a', 3] as unknown as [number, number]] },
                ],
            })
        );

        expect(model?.dock).toBeNull();
        expect(model?.forbiddenZones).toEqual([[], [[1, 2]]]);
        expect(
            toVacuumMapModel(
                buildVacuumMapPayload({ forbidden_zones: undefined, dock: { x: 'a' } as never })
            )
        ).toMatchObject({ forbiddenZones: [], dock: null });
    });
});
