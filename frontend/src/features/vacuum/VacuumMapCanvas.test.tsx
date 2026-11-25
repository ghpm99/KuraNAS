import { fireEvent, render, screen } from '@testing-library/react';
import VacuumMapCanvas from './VacuumMapCanvas';
import { buildVacuumMapPayload } from './testSupport/buildVacuumMapPayload';
import { createFakeCanvasContext } from './testSupport/createFakeCanvasContext';
import { mockCanvasGetContext } from './testSupport/mockCanvasGetContext';
import { mockElementBounds } from './testSupport/mockElementBounds';
import {
    frameHeightPixels,
    frameWidthPixels,
    tapOnRoomOne,
} from './testSupport/vacuumMapTaps';
import { toVacuumMapModel } from './vacuumMapModel';

const map = toVacuumMapModel(buildVacuumMapPayload());
if (!map) throw new Error('fixture must build a map');

const renderCanvas = (overrides: Partial<Parameters<typeof VacuumMapCanvas>[0]> = {}) => {
    const onTap = jest.fn();
    render(
        <VacuumMapCanvas
            map={map}
            selectedRoomIds={new Set()}
            pathPoints={[]}
            robotPosition={null}
            spotCenter={null}
            onTap={onTap}
            {...overrides}
        />
    );
    return onTap;
};

describe('VacuumMapCanvas', () => {
    afterEach(() => jest.restoreAllMocks());

    it('mounts when the browser gives no 2d context', () => {
        mockCanvasGetContext(null);

        expect(() => renderCanvas()).not.toThrow();
        expect(screen.getByRole('img', { name: 'VACUUM_MAP_CANVAS_LABEL' })).toBeInTheDocument();
    });

    it('keeps the map proportion and a canvas as big as the frame in device pixels', () => {
        mockCanvasGetContext(null);
        mockElementBounds(frameWidthPixels, frameHeightPixels);
        window.devicePixelRatio = 2;

        renderCanvas();

        const overlay = screen.getByRole('img', { name: 'VACUUM_MAP_CANVAS_LABEL' });
        expect(overlay).toHaveAttribute('width', '800');
        expect(overlay).toHaveAttribute('height', '600');
        expect(overlay.parentElement).toHaveStyle({ aspectRatio: '4 / 3' });
        window.devicePixelRatio = 1;
    });

    it('draws the cells and the overlay', () => {
        const context = createFakeCanvasContext();
        mockCanvasGetContext(context);
        mockElementBounds(frameWidthPixels, frameHeightPixels);

        renderCanvas({ robotPosition: { x: 25, y: 25 } });

        expect(context.putImageData).toHaveBeenCalled();
        expect(context.arc).toHaveBeenCalled();
    });

    it('does not draw the overlay before the frame has a width', () => {
        const context = createFakeCanvasContext();
        mockCanvasGetContext(context);

        renderCanvas({ robotPosition: { x: 25, y: 25 } });

        expect(context.putImageData).toHaveBeenCalled();
        expect(context.arc).not.toHaveBeenCalled();
    });

    it('reports the tapped point in millimeters, y growing upward', () => {
        mockCanvasGetContext(null);
        mockElementBounds(frameWidthPixels, frameHeightPixels);
        const onTap = renderCanvas();

        fireEvent.click(screen.getByRole('img'), tapOnRoomOne);

        expect(onTap).toHaveBeenCalledWith({ x: -25, y: 25 });
    });

    it('ignores a tap while the canvas has no size', () => {
        mockCanvasGetContext(null);
        const onTap = renderCanvas();

        fireEvent.click(screen.getByRole('img'), tapOnRoomOne);

        expect(onTap).not.toHaveBeenCalled();
    });
});
