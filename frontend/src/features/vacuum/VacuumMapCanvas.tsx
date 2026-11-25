import { useEffect, useRef, type MouseEvent } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { drawVacuumMapCells } from './drawVacuumMapCells';
import { drawVacuumMapOverlay } from './drawVacuumMapOverlay';
import styles from './VacuumMapCanvas.module.css';
import { useMeasuredElementWidth } from './useMeasuredElementWidth';
import {
    buildSpotCleaningSquare,
    mapFractionToMillimeters,
    type MillimeterPoint,
} from './vacuumMapGeometry';
import type { VacuumMapModel } from './vacuumMapModel';
import { vacuumMapPalette } from './vacuumMapPalette';

type VacuumMapCanvasProps = {
    map: VacuumMapModel;
    selectedRoomIds: ReadonlySet<number>;
    pathPoints: readonly (readonly [number, number])[];
    robotPosition: MillimeterPoint | null;
    spotCenter: MillimeterPoint | null;
    onTap: (point: MillimeterPoint) => void;
};

const VacuumMapCanvas = ({
    map,
    selectedRoomIds,
    pathPoints,
    robotPosition,
    spotCenter,
    onTap,
}: VacuumMapCanvasProps) => {
    const { t } = useI18n();
    const frameRef = useRef<HTMLDivElement>(null);
    const cellsCanvasRef = useRef<HTMLCanvasElement>(null);
    const overlayCanvasRef = useRef<HTMLCanvasElement>(null);
    const frameWidthCssPixels = useMeasuredElementWidth(frameRef);
    const devicePixelRatio = window.devicePixelRatio || 1;
    const overlayWidthPixels = Math.round(frameWidthCssPixels * devicePixelRatio);
    const overlayHeightPixels = Math.round((overlayWidthPixels * map.height) / map.width);

    useEffect(() => {
        const context = cellsCanvasRef.current?.getContext('2d');
        if (!context) return;
        drawVacuumMapCells(context, map, selectedRoomIds);
    }, [map, selectedRoomIds]);

    useEffect(() => {
        const context = overlayCanvasRef.current?.getContext('2d');
        if (!context || overlayWidthPixels === 0) return;
        drawVacuumMapOverlay(context, {
            map,
            canvasWidthPixels: overlayWidthPixels,
            canvasHeightPixels: overlayHeightPixels,
            devicePixelRatio,
            pathPoints,
            robotPosition,
            spotSquare: spotCenter === null ? null : buildSpotCleaningSquare(spotCenter),
            spotCenter,
        });
    }, [
        map,
        overlayWidthPixels,
        overlayHeightPixels,
        devicePixelRatio,
        pathPoints,
        robotPosition,
        spotCenter,
    ]);

    const handleClick = (event: MouseEvent<HTMLCanvasElement>) => {
        const bounds = event.currentTarget.getBoundingClientRect();
        if (bounds.width === 0 || bounds.height === 0) return;
        onTap(
            mapFractionToMillimeters(
                map,
                (event.clientX - bounds.left) / bounds.width,
                (event.clientY - bounds.top) / bounds.height
            )
        );
    };

    return (
        <div
            ref={frameRef}
            className={styles.frame}
            style={{
                aspectRatio: `${map.width} / ${map.height}`,
                background: vacuumMapPalette.backdrop,
            }}
        >
            <canvas
                ref={cellsCanvasRef}
                className={styles.layer}
                width={map.width}
                height={map.height}
                aria-hidden="true"
            />
            <canvas
                ref={overlayCanvasRef}
                className={`${styles.layer} ${styles.interactive}`}
                width={overlayWidthPixels}
                height={overlayHeightPixels}
                role="img"
                aria-label={t('VACUUM_MAP_CANVAS_LABEL')}
                onClick={handleClick}
            />
        </div>
    );
};

export default VacuumMapCanvas;
