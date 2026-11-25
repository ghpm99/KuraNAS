import {
    millimetersToCanvasPixels,
    type MillimeterPoint,
    type PixelPoint,
} from './vacuumMapGeometry';
import type { VacuumMapModel, VacuumMapPolygon } from './vacuumMapModel';
import { vacuumMapPalette } from './vacuumMapPalette';

export type OverlayDrawingContext = Pick<
    CanvasRenderingContext2D,
    | 'clearRect'
    | 'beginPath'
    | 'moveTo'
    | 'lineTo'
    | 'closePath'
    | 'arc'
    | 'fill'
    | 'stroke'
    | 'setLineDash'
> &
    Partial<Pick<CanvasRenderingContext2D, 'fillStyle' | 'strokeStyle' | 'lineWidth' | 'lineJoin'>>;

export type VacuumMapOverlayScene = {
    map: VacuumMapModel;
    canvasWidthPixels: number;
    canvasHeightPixels: number;
    devicePixelRatio: number;
    pathPoints: readonly (readonly [number, number])[];
    robotPosition: MillimeterPoint | null;
    spotSquare: VacuumMapPolygon | null;
    spotCenter: MillimeterPoint | null;
};

const markerRadiusCssPixels = 7;
const pathWidthCssPixels = 2;
const outlineWidthCssPixels = 2;
const dashLengthCssPixels = 6;
const fullCircleRadians = Math.PI * 2;

const toPixels = (scene: VacuumMapOverlayScene, point: MillimeterPoint): PixelPoint =>
    millimetersToCanvasPixels(scene.map, point, scene.canvasWidthPixels / scene.map.width);

const tracePolygon = (
    context: OverlayDrawingContext,
    scene: VacuumMapOverlayScene,
    polygon: VacuumMapPolygon
): void => {
    context.beginPath();
    polygon.forEach(([x, y], vertexIndex) => {
        const pixel = toPixels(scene, { x, y });
        if (vertexIndex === 0) context.moveTo(pixel.x, pixel.y);
        else context.lineTo(pixel.x, pixel.y);
    });
    context.closePath();
};

const drawForbiddenZones = (context: OverlayDrawingContext, scene: VacuumMapOverlayScene): void => {
    context.fillStyle = vacuumMapPalette.forbiddenFill;
    context.strokeStyle = vacuumMapPalette.forbiddenStroke;
    context.lineWidth = outlineWidthCssPixels * scene.devicePixelRatio;
    scene.map.forbiddenZones.forEach((zone) => {
        if (zone.length < 3) return;
        tracePolygon(context, scene, zone);
        context.fill();
        context.stroke();
    });
};

const drawPath = (context: OverlayDrawingContext, scene: VacuumMapOverlayScene): void => {
    if (scene.pathPoints.length < 2) return;
    context.strokeStyle = vacuumMapPalette.path;
    context.lineWidth = pathWidthCssPixels * scene.devicePixelRatio;
    context.lineJoin = 'round';
    context.beginPath();
    scene.pathPoints.forEach(([x, y], pointIndex) => {
        const pixel = toPixels(scene, { x, y });
        if (pointIndex === 0) context.moveTo(pixel.x, pixel.y);
        else context.lineTo(pixel.x, pixel.y);
    });
    context.stroke();
};

const drawMarker = (
    context: OverlayDrawingContext,
    scene: VacuumMapOverlayScene,
    position: MillimeterPoint,
    fillColor: string,
    ringColor: string
): void => {
    const pixel = toPixels(scene, position);
    const radius = markerRadiusCssPixels * scene.devicePixelRatio;
    context.beginPath();
    context.arc(pixel.x, pixel.y, radius, 0, fullCircleRadians);
    context.fillStyle = fillColor;
    context.fill();
    context.strokeStyle = ringColor;
    context.lineWidth = outlineWidthCssPixels * scene.devicePixelRatio;
    context.stroke();
};

const drawSpot = (context: OverlayDrawingContext, scene: VacuumMapOverlayScene): void => {
    if (scene.spotSquare === null || scene.spotCenter === null) return;
    context.fillStyle = vacuumMapPalette.spotFill;
    context.strokeStyle = vacuumMapPalette.spotStroke;
    context.lineWidth = outlineWidthCssPixels * scene.devicePixelRatio;
    tracePolygon(context, scene, scene.spotSquare);
    context.fill();
    context.setLineDash([
        dashLengthCssPixels * scene.devicePixelRatio,
        dashLengthCssPixels * scene.devicePixelRatio,
    ]);
    context.stroke();
    context.setLineDash([]);
    drawMarker(context, scene, scene.spotCenter, vacuumMapPalette.spotStroke, vacuumMapPalette.robotRing);
};

/** Draws zones, path, "clean here" square, dock and robot over the cells, in canvas pixels. */
export const drawVacuumMapOverlay = (
    context: OverlayDrawingContext,
    scene: VacuumMapOverlayScene
): void => {
    context.clearRect(0, 0, scene.canvasWidthPixels, scene.canvasHeightPixels);
    drawForbiddenZones(context, scene);
    drawPath(context, scene);
    drawSpot(context, scene);
    if (scene.map.dock) {
        drawMarker(context, scene, scene.map.dock, vacuumMapPalette.dock, vacuumMapPalette.robotRing);
    }
    if (scene.robotPosition) {
        drawMarker(
            context,
            scene,
            scene.robotPosition,
            vacuumMapPalette.robot,
            vacuumMapPalette.robotRing
        );
    }
};
