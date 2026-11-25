import type { VacuumMapModel } from './vacuumMapModel';
import { isRoomFloorCellValue } from './vacuumMapCellValues';

export type MillimeterPoint = { x: number; y: number };
export type PixelPoint = { x: number; y: number };

export const spotCleaningSideMillimeters = 1500;

type MapGrid = Pick<VacuumMapModel, 'width' | 'height' | 'resolution' | 'originX' | 'originY'>;

/** Converts a point in millimeters to canvas pixels; the map's y axis grows upward. */
export const millimetersToCanvasPixels = (
    grid: MapGrid,
    point: MillimeterPoint,
    pixelsPerCell: number
): PixelPoint => ({
    x: ((point.x - grid.originX) / grid.resolution) * pixelsPerCell,
    y: (grid.height - (point.y - grid.originY) / grid.resolution) * pixelsPerCell,
});

/** Converts a position inside the drawn map (0..1 from the top-left corner) to millimeters. */
export const mapFractionToMillimeters = (
    grid: MapGrid,
    fractionFromLeft: number,
    fractionFromTop: number
): MillimeterPoint => ({
    x: grid.originX + fractionFromLeft * grid.width * grid.resolution,
    y: grid.originY + (1 - fractionFromTop) * grid.height * grid.resolution,
});

/** Index into the cells array for the cell under a point, or -1 when the point is outside the grid. */
export const cellIndexAt = (grid: MapGrid, point: MillimeterPoint): number => {
    const column = Math.floor((point.x - grid.originX) / grid.resolution);
    const row = Math.floor((point.y - grid.originY) / grid.resolution);
    if (column < 0 || column >= grid.width || row < 0 || row >= grid.height) return -1;
    return row * grid.width + column;
};

/** Id of the room whose floor is under the point, or null for walls, obstacles, unknown and outside. */
export const findRoomIdAt = (map: VacuumMapModel, point: MillimeterPoint): number | null => {
    const cellIndex = cellIndexAt(map, point);
    if (cellIndex < 0) return null;
    const cellValue = map.cells[cellIndex] ?? 0;
    return isRoomFloorCellValue(cellValue) ? cellValue : null;
};

/** Corners of the square cleaned around a point by "clean here". */
export const buildSpotCleaningSquare = (center: MillimeterPoint): [number, number][] => {
    const halfSide = spotCleaningSideMillimeters / 2;
    return [
        [center.x - halfSide, center.y - halfSide],
        [center.x + halfSide, center.y - halfSide],
        [center.x + halfSide, center.y + halfSide],
        [center.x - halfSide, center.y + halfSide],
    ];
};
