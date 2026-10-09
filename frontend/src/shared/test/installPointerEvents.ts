type PointerEventInit = MouseEventInit & { pointerId?: number; pointerType?: string };

class PointerEventWithCoordinates extends MouseEvent {
    readonly pointerId: number;
    readonly pointerType: string;

    constructor(type: string, init: PointerEventInit = {}) {
        super(type, init);
        this.pointerId = init.pointerId ?? 1;
        this.pointerType = init.pointerType ?? '';
    }
}

export const installPointerEvents = () => {
    const originalPointerEvent = window.PointerEvent;
    window.PointerEvent = PointerEventWithCoordinates as unknown as typeof PointerEvent;
    return () => {
        window.PointerEvent = originalPointerEvent;
    };
};
