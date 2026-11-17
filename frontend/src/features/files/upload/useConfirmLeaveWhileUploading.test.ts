import { renderHook } from '@testing-library/react';
import useConfirmLeaveWhileUploading from './useConfirmLeaveWhileUploading';

const dispatchBeforeUnload = (): BeforeUnloadEvent => {
    const event = new Event('beforeunload', { cancelable: true }) as BeforeUnloadEvent;
    Object.defineProperty(event, 'returnValue', { value: 'untouched', writable: true });
    window.dispatchEvent(event);
    return event;
};

describe('useConfirmLeaveWhileUploading', () => {
    it('does not block leaving the page without active uploads', () => {
        renderHook(() => useConfirmLeaveWhileUploading(false));

        const event = dispatchBeforeUnload();

        expect(event.defaultPrevented).toBe(false);
        expect(event.returnValue).toBe('untouched');
    });

    it('asks the browser to confirm leaving while uploads are active', () => {
        renderHook(() => useConfirmLeaveWhileUploading(true));

        const event = dispatchBeforeUnload();

        expect(event.defaultPrevented).toBe(true);
        expect(event.returnValue).toBe('');
    });

    it('stops blocking when uploads finish', () => {
        const { rerender } = renderHook(
            ({ hasActiveUploads }) => useConfirmLeaveWhileUploading(hasActiveUploads),
            { initialProps: { hasActiveUploads: true } }
        );
        rerender({ hasActiveUploads: false });

        expect(dispatchBeforeUnload().defaultPrevented).toBe(false);
    });

    it('stops blocking when unmounted', () => {
        const { unmount } = renderHook(() => useConfirmLeaveWhileUploading(true));
        unmount();

        expect(dispatchBeforeUnload().defaultPrevented).toBe(false);
    });
});
