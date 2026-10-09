export type ArrowDirection = 'up' | 'down' | 'left' | 'right';

type ShortcutAction =
    | { kind: 'openItem' }
    | { kind: 'deleteSelection' }
    | { kind: 'renameSelection' }
    | { kind: 'selectAll' }
    | { kind: 'goToParent' }
    | { kind: 'focusSearch' }
    | { kind: 'moveFocus'; direction: ArrowDirection; isExtendingSelection: boolean };

type KeyboardEventKeys = Pick<KeyboardEvent, 'key' | 'ctrlKey' | 'metaKey' | 'altKey' | 'shiftKey'>;

const arrowDirectionsByKey: Record<string, ArrowDirection> = {
    ArrowUp: 'up',
    ArrowDown: 'down',
    ArrowLeft: 'left',
    ArrowRight: 'right',
};

const plainKeyActions: Record<string, ShortcutAction> = {
    Enter: { kind: 'openItem' },
    Delete: { kind: 'deleteSelection' },
    F2: { kind: 'renameSelection' },
    Backspace: { kind: 'goToParent' },
    '/': { kind: 'focusSearch' },
};

export const resolveShortcutAction = (event: KeyboardEventKeys): ShortcutAction | null => {
    const isCommandKey = event.ctrlKey || event.metaKey;
    if (isCommandKey) {
        const isSelectAll = event.key.toLowerCase() === 'a' && !event.altKey && !event.shiftKey;
        return isSelectAll ? { kind: 'selectAll' } : null;
    }
    if (event.altKey) {
        return event.key === 'ArrowUp' && !event.shiftKey ? { kind: 'goToParent' } : null;
    }
    const direction = arrowDirectionsByKey[event.key];
    if (direction) {
        return { kind: 'moveFocus', direction, isExtendingSelection: event.shiftKey };
    }
    if (event.shiftKey) return null;
    return plainKeyActions[event.key] ?? null;
};

export const resolveNextFocusIndex = ({
    currentIndex,
    direction,
    columnCount,
    itemCount,
}: {
    currentIndex: number;
    direction: ArrowDirection;
    columnCount: number;
    itemCount: number;
}): number => {
    if (itemCount === 0) return -1;
    if (currentIndex < 0) return 0;
    const stepByDirection: Record<ArrowDirection, number> = {
        left: -1,
        right: 1,
        up: -columnCount,
        down: columnCount,
    };
    const nextIndex = currentIndex + stepByDirection[direction];
    return Math.min(Math.max(nextIndex, 0), itemCount - 1);
};
