import { buildVacuumMapPixels } from './buildVacuumMapPixels';
import type { VacuumMapModel } from './vacuumMapModel';

export type CellsDrawingContext = Pick<
    CanvasRenderingContext2D,
    'createImageData' | 'putImageData' | 'clearRect'
>;

/** Paints the grid at one pixel per cell; the canvas element scales it up. */
export const drawVacuumMapCells = (
    context: CellsDrawingContext,
    map: VacuumMapModel,
    selectedRoomIds: ReadonlySet<number>
): void => {
    context.clearRect(0, 0, map.width, map.height);
    const image = context.createImageData(map.width, map.height);
    image.data.set(buildVacuumMapPixels(map, selectedRoomIds));
    context.putImageData(image, 0, 0);
};
