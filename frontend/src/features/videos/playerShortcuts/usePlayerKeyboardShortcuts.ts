import { useEffect, useRef, useState } from 'react';
import {
    runPlayerShortcut,
    type PlayerShortcutControls,
    type PlayerShortcutFeedback,
    type PlayerShortcutSnapshot,
} from './playerShortcutExecution';
import { classifyShortcutTarget, resolvePlayerShortcut } from './playerShortcutMapping';

export const SHORTCUT_FEEDBACK_DURATION_MS = 600;

const openOverlaySelector = '[role="dialog"], [role="alertdialog"], [role="menu"]';

interface UsePlayerKeyboardShortcutsOptions {
    snapshot: PlayerShortcutSnapshot;
    controls: PlayerShortcutControls;
    isSuspended: boolean;
}

export interface ShownShortcutFeedback extends PlayerShortcutFeedback {
    id: number;
}

const isOverlayOpenInDocument = () => document.querySelector(openOverlaySelector) !== null;

export default function usePlayerKeyboardShortcuts({
    snapshot,
    controls,
    isSuspended,
}: UsePlayerKeyboardShortcutsOptions) {
    const [feedback, setFeedback] = useState<ShownShortcutFeedback | null>(null);
    const latestRef = useRef({ snapshot, controls, isSuspended });
    const feedbackCounterRef = useRef(0);

    useEffect(() => {
        latestRef.current = { snapshot, controls, isSuspended };
    });

    useEffect(() => {
        const handleKeyDown = (event: KeyboardEvent) => {
            const latest = latestRef.current;
            if (latest.isSuspended || isOverlayOpenInDocument()) return;

            const action = resolvePlayerShortcut({
                key: event.key,
                shiftKey: event.shiftKey,
                ctrlKey: event.ctrlKey,
                metaKey: event.metaKey,
                altKey: event.altKey,
                targetKind: classifyShortcutTarget(event.target),
            });
            if (!action) return;

            event.preventDefault();
            const shownFeedback = runPlayerShortcut(action, latest.snapshot, latest.controls);
            if (!shownFeedback) return;
            feedbackCounterRef.current += 1;
            setFeedback({ ...shownFeedback, id: feedbackCounterRef.current });
        };

        document.addEventListener('keydown', handleKeyDown);
        return () => document.removeEventListener('keydown', handleKeyDown);
    }, []);

    useEffect(() => {
        if (!feedback) return;
        const timeoutId = setTimeout(() => setFeedback(null), SHORTCUT_FEEDBACK_DURATION_MS);
        return () => clearTimeout(timeoutId);
    }, [feedback]);

    return { feedback };
}
