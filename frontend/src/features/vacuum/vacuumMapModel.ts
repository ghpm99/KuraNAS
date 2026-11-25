import type { VacuumMap } from '@/types/vacuum';
import { decodeVacuumMapCells } from './decodeVacuumMapCells';

export type VacuumMapPolygon = readonly (readonly [number, number])[];

export type VacuumMapModel = {
    mapId: number;
    width: number;
    height: number;
    resolution: number;
    originX: number;
    originY: number;
    cells: Uint8Array;
    dock: { x: number; y: number } | null;
    forbiddenZones: readonly VacuumMapPolygon[];
};

const isPositiveNumber = (candidate: unknown): candidate is number =>
    typeof candidate === 'number' && Number.isFinite(candidate) && candidate > 0;

const toFiniteNumber = (candidate: unknown): number =>
    typeof candidate === 'number' && Number.isFinite(candidate) ? candidate : 0;

const toPolygon = (vertices: unknown): VacuumMapPolygon => {
    if (!Array.isArray(vertices)) return [];
    return vertices
        .filter(
            (vertex): vertex is [number, number] =>
                Array.isArray(vertex) &&
                typeof vertex[0] === 'number' &&
                typeof vertex[1] === 'number'
        )
        .map(([x, y]) => [x, y] as const);
};

const toDock = (dock: Partial<VacuumMap>['dock']): VacuumMapModel['dock'] => {
    if (!dock || typeof dock.x !== 'number' || typeof dock.y !== 'number') return null;
    return { x: dock.x, y: dock.y };
};

/**
 * Turns the server payload into a drawable map. Returns null when the
 * payload is partial or its grid does not have width * height cells.
 */
export const toVacuumMapModel = (map: Partial<VacuumMap> | undefined): VacuumMapModel | null => {
    if (!map) return null;
    if (!isPositiveNumber(map.width) || !isPositiveNumber(map.height)) return null;
    if (!isPositiveNumber(map.resolution)) return null;
    const cells = decodeVacuumMapCells(map.cells);
    if (cells.length !== map.width * map.height) return null;
    return {
        mapId: toFiniteNumber(map.map_id),
        width: map.width,
        height: map.height,
        resolution: map.resolution,
        originX: toFiniteNumber(map.origin_x),
        originY: toFiniteNumber(map.origin_y),
        cells,
        dock: toDock(map.dock),
        forbiddenZones: Array.isArray(map.forbidden_zones)
            ? map.forbidden_zones.map((zone) => toPolygon(zone?.vertices))
            : [],
    };
};
