import { act, fireEvent, render, screen } from '@testing-library/react';
import { installPointerEvents } from '@/shared/test/installPointerEvents';
import { useImageViewerGestures } from './useImageViewerGestures';

type GestureHarnessProps = {
    zoom?: number;
    pan?: { x: number; y: number };
    onZoomChange?: (zoom: number) => void;
    onPanChange?: (panX: number, panY: number) => void;
    onPrevious?: () => void;
    onNext?: () => void;
};

const noPan = { x: 0, y: 0 };

function GestureHarness({
    zoom = 1,
    pan = noPan,
    onZoomChange = jest.fn(),
    onPanChange = jest.fn(),
    onPrevious = jest.fn(),
    onNext = jest.fn(),
}: GestureHarnessProps) {
    const handlers = useImageViewerGestures({
        zoom,
        pan,
        onZoomChange,
        onPanChange,
        onPrevious,
        onNext,
    });
    return <div data-testid="surface" {...handlers} />;
}

const surface = () => screen.getByTestId('surface');

const touchPointer = (pointerId: number, clientX: number, clientY: number) => ({
    pointerId,
    pointerType: 'touch',
    clientX,
    clientY,
});

const swipe = (fromX: number, toX: number, verticalDrift = 0) => {
    fireEvent.pointerDown(surface(), touchPointer(1, fromX, 100));
    fireEvent.pointerMove(surface(), touchPointer(1, (fromX + toX) / 2, 100));
    fireEvent.pointerUp(surface(), touchPointer(1, toX, 100 + verticalDrift));
};

describe('useImageViewerGestures', () => {
    let restorePointerEvents: () => void;

    beforeEach(() => {
        restorePointerEvents = installPointerEvents();
        jest.useFakeTimers();
        jest.setSystemTime(new Date('2026-01-01T00:00:00Z'));
    });

    afterEach(() => {
        restorePointerEvents();
        jest.useRealTimers();
    });

    it('renders its surface and ignores moves from pointers it never saw', () => {
        const onPanChange = jest.fn();
        render(<GestureHarness zoom={2} onPanChange={onPanChange} />);

        fireEvent.pointerMove(surface(), touchPointer(9, 50, 50));
        fireEvent.pointerUp(surface(), touchPointer(9, 50, 50));

        expect(surface()).toBeInTheDocument();
        expect(onPanChange).not.toHaveBeenCalled();
    });

    it('goes to the next image on a left swipe and to the previous on a right swipe', () => {
        const onNext = jest.fn();
        const onPrevious = jest.fn();
        render(<GestureHarness onNext={onNext} onPrevious={onPrevious} />);

        swipe(300, 100);
        expect(onNext).toHaveBeenCalledTimes(1);
        expect(onPrevious).not.toHaveBeenCalled();

        swipe(100, 300);
        expect(onPrevious).toHaveBeenCalledTimes(1);
        expect(onNext).toHaveBeenCalledTimes(1);
    });

    it('ignores short swipes, mostly vertical swipes and cancelled gestures', () => {
        const onNext = jest.fn();
        const onPrevious = jest.fn();
        render(<GestureHarness onNext={onNext} onPrevious={onPrevious} />);

        swipe(200, 170);
        swipe(300, 200, 120);
        fireEvent.pointerDown(surface(), touchPointer(1, 300, 100));
        fireEvent.pointerCancel(surface(), touchPointer(1, 100, 100));

        expect(onNext).not.toHaveBeenCalled();
        expect(onPrevious).not.toHaveBeenCalled();
    });

    it('does not swipe between images while zoomed in, it pans instead', () => {
        const onNext = jest.fn();
        const onPanChange = jest.fn();
        render(
            <GestureHarness
                zoom={2}
                pan={{ x: 10, y: 5 }}
                onNext={onNext}
                onPanChange={onPanChange}
            />
        );

        fireEvent.pointerDown(surface(), touchPointer(1, 300, 100));
        fireEvent.pointerMove(surface(), touchPointer(1, 250, 130));
        fireEvent.pointerUp(surface(), touchPointer(1, 100, 100));

        expect(onNext).not.toHaveBeenCalled();
        expect(onPanChange).toHaveBeenCalledWith(-40, 35);
    });

    it('pans with a mouse drag when zoomed and does not pan at 1x', () => {
        const onPanChange = jest.fn();
        const { rerender } = render(<GestureHarness zoom={1} onPanChange={onPanChange} />);

        fireEvent.pointerDown(surface(), {
            pointerId: 1,
            pointerType: 'mouse',
            clientX: 10,
            clientY: 10,
        });
        fireEvent.pointerMove(surface(), {
            pointerId: 1,
            pointerType: 'mouse',
            clientX: 40,
            clientY: 10,
        });
        fireEvent.pointerUp(surface(), {
            pointerId: 1,
            pointerType: 'mouse',
            clientX: 40,
            clientY: 10,
        });
        expect(onPanChange).not.toHaveBeenCalled();

        rerender(<GestureHarness zoom={3} onPanChange={onPanChange} />);
        fireEvent.pointerDown(surface(), {
            pointerId: 2,
            pointerType: 'mouse',
            clientX: 10,
            clientY: 10,
        });
        fireEvent.pointerMove(surface(), {
            pointerId: 2,
            pointerType: 'mouse',
            clientX: 40,
            clientY: 25,
        });
        expect(onPanChange).toHaveBeenLastCalledWith(30, 15);
    });

    it('keeps the pan inside the frame when the frame has a size', () => {
        const onPanChange = jest.fn();
        render(<GestureHarness zoom={2} onPanChange={onPanChange} />);
        jest.spyOn(surface(), 'getBoundingClientRect').mockReturnValue({
            width: 100,
            height: 100,
        } as DOMRect);

        fireEvent.pointerDown(surface(), touchPointer(1, 0, 0));
        fireEvent.pointerMove(surface(), touchPointer(1, 900, -900));

        expect(onPanChange).toHaveBeenLastCalledWith(100, -100);
    });

    it('zooms with a two finger pinch relative to the starting zoom', () => {
        const onZoomChange = jest.fn();
        render(<GestureHarness zoom={1.5} onZoomChange={onZoomChange} />);

        fireEvent.pointerDown(surface(), touchPointer(1, 100, 100));
        fireEvent.pointerDown(surface(), touchPointer(2, 200, 100));
        fireEvent.pointerMove(surface(), touchPointer(2, 300, 100));

        expect(onZoomChange).toHaveBeenLastCalledWith(3);

        fireEvent.pointerMove(surface(), touchPointer(2, 150, 100));
        expect(onZoomChange).toHaveBeenLastCalledWith(0.75);
    });

    it('does not swipe when a pinch ended with one finger lifted', () => {
        const onNext = jest.fn();
        render(<GestureHarness onNext={onNext} />);

        fireEvent.pointerDown(surface(), touchPointer(1, 300, 100));
        fireEvent.pointerDown(surface(), touchPointer(2, 400, 100));
        fireEvent.pointerUp(surface(), touchPointer(2, 400, 100));
        fireEvent.pointerUp(surface(), touchPointer(1, 100, 100));

        expect(onNext).not.toHaveBeenCalled();
    });

    it('toggles between zoomed in and 1x on a touch double tap', () => {
        const onZoomChange = jest.fn();
        const { rerender } = render(<GestureHarness zoom={1} onZoomChange={onZoomChange} />);

        fireEvent.pointerDown(surface(), touchPointer(1, 100, 100));
        fireEvent.pointerUp(surface(), touchPointer(1, 100, 100));
        act(() => {
            jest.advanceTimersByTime(120);
        });
        fireEvent.pointerDown(surface(), touchPointer(2, 104, 102));
        expect(onZoomChange).toHaveBeenCalledWith(2);

        rerender(<GestureHarness zoom={2} onZoomChange={onZoomChange} />);
        fireEvent.pointerUp(surface(), touchPointer(2, 104, 102));
        act(() => {
            jest.advanceTimersByTime(1000);
        });
        fireEvent.pointerDown(surface(), touchPointer(3, 100, 100));
        fireEvent.pointerUp(surface(), touchPointer(3, 100, 100));
        act(() => {
            jest.advanceTimersByTime(100);
        });
        fireEvent.pointerDown(surface(), touchPointer(4, 100, 100));
        expect(onZoomChange).toHaveBeenLastCalledWith(1);
    });

    it('does not treat slow or distant taps as a double tap', () => {
        const onZoomChange = jest.fn();
        render(<GestureHarness onZoomChange={onZoomChange} />);

        fireEvent.pointerDown(surface(), touchPointer(1, 100, 100));
        fireEvent.pointerUp(surface(), touchPointer(1, 100, 100));
        act(() => {
            jest.advanceTimersByTime(600);
        });
        fireEvent.pointerDown(surface(), touchPointer(2, 100, 100));
        fireEvent.pointerUp(surface(), touchPointer(2, 100, 100));
        act(() => {
            jest.advanceTimersByTime(100);
        });
        fireEvent.pointerDown(surface(), touchPointer(3, 400, 400));

        expect(onZoomChange).not.toHaveBeenCalled();
    });

    it('toggles zoom on a mouse double click but not right after a touch double tap', () => {
        const onZoomChange = jest.fn();
        render(<GestureHarness zoom={1} onZoomChange={onZoomChange} />);

        fireEvent.doubleClick(surface());
        expect(onZoomChange).toHaveBeenCalledTimes(1);
        expect(onZoomChange).toHaveBeenLastCalledWith(2);

        act(() => {
            jest.advanceTimersByTime(1000);
        });
        fireEvent.pointerDown(surface(), touchPointer(1, 100, 100));
        fireEvent.pointerUp(surface(), touchPointer(1, 100, 100));
        fireEvent.pointerDown(surface(), touchPointer(2, 100, 100));
        expect(onZoomChange).toHaveBeenCalledTimes(2);
        fireEvent.doubleClick(surface());
        expect(onZoomChange).toHaveBeenCalledTimes(2);
    });
});
