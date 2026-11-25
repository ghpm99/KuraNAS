import { vacuumMapPalette } from './vacuumMapPalette';
import {
    isRoomFloorCellValue,
    obstacleCellValue,
    wallCellValue,
} from './vacuumMapCellValues';
import type { VacuumMapModel } from './vacuumMapModel';
import { vacuumRoomColor, type RgbColor } from './vacuumRoomColor';

const channelsPerPixel = 4;
const opaqueAlpha = 255;

const colorForCell = (
    cellValue: number,
    selectedRoomIds: ReadonlySet<number>
): RgbColor | null => {
    if (cellValue === wallCellValue) return vacuumMapPalette.wall;
    if (cellValue === obstacleCellValue) return vacuumMapPalette.obstacle;
    if (isRoomFloorCellValue(cellValue)) {
        return vacuumRoomColor(cellValue, selectedRoomIds.has(cellValue));
    }
    return null;
};

/**
 * RGBA pixels of the grid, one pixel per cell, with the lowest-y row at the bottom.
 * Unknown cells stay transparent.
 */
export const buildVacuumMapPixels = (
    map: VacuumMapModel,
    selectedRoomIds: ReadonlySet<number>
): Uint8ClampedArray => {
    const pixels = new Uint8ClampedArray(map.width * map.height * channelsPerPixel);
    for (let row = 0; row < map.height; row += 1) {
        const pixelRow = map.height - 1 - row;
        for (let column = 0; column < map.width; column += 1) {
            const color = colorForCell(map.cells[row * map.width + column] ?? 0, selectedRoomIds);
            if (color === null) continue;
            const pixelOffset = (pixelRow * map.width + column) * channelsPerPixel;
            pixels[pixelOffset] = color[0];
            pixels[pixelOffset + 1] = color[1];
            pixels[pixelOffset + 2] = color[2];
            pixels[pixelOffset + 3] = opaqueAlpha;
        }
    }
    return pixels;
};
