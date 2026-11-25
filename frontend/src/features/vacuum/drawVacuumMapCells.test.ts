import { drawVacuumMapCells } from './drawVacuumMapCells';
import { createFakeCanvasContext } from './testSupport/createFakeCanvasContext';
import { buildVacuumMapPayload } from './testSupport/buildVacuumMapPayload';
import { toVacuumMapModel } from './vacuumMapModel';

const map = toVacuumMapModel(buildVacuumMapPayload());
if (!map) throw new Error('fixture must build a map');

describe('drawVacuumMapCells', () => {
    it('clears the canvas and puts one pixel per cell', () => {
        const context = createFakeCanvasContext();

        drawVacuumMapCells(context as never, map, new Set());

        expect(context.clearRect).toHaveBeenCalledWith(0, 0, 4, 3);
        expect(context.createImageData).toHaveBeenCalledWith(4, 3);
        expect(context.imageData.data).toHaveLength(48);
        expect(context.putImageData).toHaveBeenCalledWith(context.imageData, 0, 0);
    });
});
