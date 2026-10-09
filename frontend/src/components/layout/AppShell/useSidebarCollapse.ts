import useMediaQuery from '@mui/material/useMediaQuery';
import { useCallback, useEffect, useState } from 'react';
import { subscribeToSidebarToggle } from '../appCommandEvents';
import {
    compactDesktopMediaQuery,
    isTypingTarget,
    readSidebarCollapsedChoice,
    sidebarWidthCssVariable,
    writeSidebarCollapsedChoice,
} from './sidebarCollapsePreference';

const TOGGLE_SHORTCUT_KEY = '[';
const COLLAPSED_WIDTH_VARIABLE = 'var(--app-shell-sidebar-width-collapsed)';
const EXPANDED_WIDTH_VARIABLE = 'var(--app-shell-sidebar-width)';

export const useSidebarCollapse = () => {
    const isCompactDesktop = useMediaQuery(compactDesktopMediaQuery);
    const [userChoice, setUserChoice] = useState<boolean | null>(readSidebarCollapsedChoice);
    const isCollapsed = userChoice ?? isCompactDesktop;

    const toggleCollapsed = useCallback(() => {
        const nextCollapsed = !isCollapsed;
        writeSidebarCollapsedChoice(nextCollapsed);
        setUserChoice(nextCollapsed);
    }, [isCollapsed]);

    useEffect(() => {
        const handleKeyDown = (event: KeyboardEvent) => {
            if (event.key !== TOGGLE_SHORTCUT_KEY) return;
            if (event.ctrlKey || event.metaKey || event.altKey) return;
            if (isTypingTarget(event.target)) return;
            toggleCollapsed();
        };
        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [toggleCollapsed]);

    useEffect(() => subscribeToSidebarToggle(toggleCollapsed), [toggleCollapsed]);

    useEffect(() => {
        const rootElement = document.documentElement;
        rootElement.style.setProperty(
            sidebarWidthCssVariable,
            isCollapsed ? COLLAPSED_WIDTH_VARIABLE : EXPANDED_WIDTH_VARIABLE
        );
        return () => {
            rootElement.style.removeProperty(sidebarWidthCssVariable);
        };
    }, [isCollapsed]);

    return { isCollapsed, toggleCollapsed };
};
