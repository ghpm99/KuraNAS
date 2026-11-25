import { drawVacuumMapOverlay, type VacuumMapOverlayScene } from './drawVacuumMapOverlay';
import { createFakeCanvasContext } from './testSupport/createFakeCanvasContext';
import { buildVacuumMapPayload } from './testSupport/buildVacuumMapPayload';
import { vacuumMapPalette } from './vacuumMapPalette';
import { toVacuumMapModel } from './vacuumMapModel';

const map = toVacuumMapModel(buildVacuumMapPayload());
if (!map) throw new Error('fixture must build a map');

const emptyScene: VacuumMapOverlayScene = {
    map: { ...map, dock: null, forbiddenZones: [] },
    canvasWidthPixels: 40,
    canvasHeightPixels: 30,
    devicePixelRatio: 2,
    pathPoints: [],
    robotPosition: null,
    spotSquare: null,
    spotCenter: null,
};

describe('drawVacuumMapOverlay', () => {
    it('only clears the canvas when there is nothing to draw', () => {
        const context = createFakeCanvasContext();

        drawVacuumMapOverlay(context, emptyScene);

        expect(context.clearRect).toHaveBeenCalledWith(0, 0, 40, 30);
        expect(context.beginPath).not.toHaveBeenCalled();
    });

    it('fills and outlines each forbidden zone with a polygon in pixels', () => {
        const context = createFakeCanvasContext();

        drawVacuumMapOverlay(context, {
            ...emptyScene,
            map: { ...map, dock: null },
        });

        expect(context.moveTo).toHaveBeenCalledWith(20, 20);
        expect(context.lineTo).toHaveBeenCalledWith(30, 20);
        expect(context.lineTo).toHaveBeenCalledWith(30, 10);
        expect(context.fill).toHaveBeenCalledTimes(1);
        expect(context.stroke).toHaveBeenCalledTimes(1);
        expect(context.strokeStyle).toBe(vacuumMapPalette.forbiddenStroke);
    });

    it('skips zones with fewer than three vertices', () => {
        const context = createFakeCanvasContext();

        drawVacuumMapOverlay(context, {
            ...emptyScene,
            map: {
                ...map,
                dock: null,
                forbiddenZones: [
                    [
                        [0, 0],
                        [1, 1],
                    ],
                ],
            },
        });

        expect(context.fill).not.toHaveBeenCalled();
    });

    it('draws the path as one polyline and ignores a single point', () => {
        const singlePointContext = createFakeCanvasContext();
        drawVacuumMapOverlay(singlePointContext, { ...emptyScene, pathPoints: [[0, 0]] });
        expect(singlePointContext.stroke).not.toHaveBeenCalled();

        const context = createFakeCanvasContext();
        drawVacuumMapOverlay(context, {
            ...emptyScene,
            pathPoints: [
                [0, 0],
                [50, 50],
            ],
        });

        expect(context.moveTo).toHaveBeenCalledWith(20, 20);
        expect(context.lineTo).toHaveBeenCalledWith(30, 10);
        expect(context.stroke).toHaveBeenCalledTimes(1);
        expect(context.lineWidth).toBe(4);
    });

    it('marks the dock and the robot', () => {
        const context = createFakeCanvasContext();

        drawVacuumMapOverlay(context, {
            ...emptyScene,
            map: { ...map, forbiddenZones: [] },
            robotPosition: { x: 75, y: 25 },
        });

        expect(context.arc).toHaveBeenCalledTimes(2);
        expect(context.arc).toHaveBeenCalledWith(25, 15, 14, 0, Math.PI * 2);
        expect(context.arc).toHaveBeenCalledWith(35, 15, 14, 0, Math.PI * 2);
    });

    it('draws the clean here square dashed with its center marker', () => {
        const context = createFakeCanvasContext();

        drawVacuumMapOverlay(context, {
            ...emptyScene,
            spotSquare: [
                [0, 0],
                [50, 0],
                [50, 50],
                [0, 50],
            ],
            spotCenter: { x: 25, y: 25 },
        });

        expect(context.setLineDash).toHaveBeenNthCalledWith(1, [12, 12]);
        expect(context.setLineDash).toHaveBeenLastCalledWith([]);
        expect(context.arc).toHaveBeenCalledTimes(1);
    });
});
