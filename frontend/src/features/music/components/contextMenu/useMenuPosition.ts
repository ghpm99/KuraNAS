import { useCallback, useState } from 'react';
import type { MouseEvent } from 'react';

export type MenuPosition = { top: number; left: number };

export default function useMenuPosition() {
    const [position, setPosition] = useState<MenuPosition | null>(null);

    const openFromButton = useCallback((event: MouseEvent<HTMLElement>) => {
        event.stopPropagation();
        const buttonBounds = event.currentTarget.getBoundingClientRect();
        setPosition({ top: buttonBounds.bottom, left: buttonBounds.left });
    }, []);

    const openFromContextMenuEvent = useCallback((event: MouseEvent<HTMLElement>) => {
        event.preventDefault();
        event.stopPropagation();
        setPosition({ top: event.clientY, left: event.clientX });
    }, []);

    const close = useCallback(() => setPosition(null), []);

    return { position, openFromButton, openFromContextMenuEvent, close };
}
