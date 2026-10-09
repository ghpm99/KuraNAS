import { useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { appRoutes } from '@/app/routes';
import { isModalOpen, isTextEntryTarget } from './keyboardEventContext';

const GO_SEQUENCE_KEY = 'g';
const HELP_KEY = '?';
const GO_SEQUENCE_TIMEOUT_MS = 1500;

export const routesByGoSequenceKey: Record<string, string> = {
    h: appRoutes.home,
    f: appRoutes.files,
    i: appRoutes.images,
    m: appRoutes.music,
    v: appRoutes.videos,
    s: appRoutes.settings,
};

export const useGlobalShortcuts = (onShowHelp: () => void) => {
    const navigate = useNavigate();
    const isAwaitingGoTargetRef = useRef(false);
    const goSequenceTimeoutRef = useRef<number | undefined>(undefined);

    useEffect(() => {
        const cancelGoSequence = () => {
            isAwaitingGoTargetRef.current = false;
            window.clearTimeout(goSequenceTimeoutRef.current);
        };

        const startGoSequence = () => {
            cancelGoSequence();
            isAwaitingGoTargetRef.current = true;
            goSequenceTimeoutRef.current = window.setTimeout(cancelGoSequence, GO_SEQUENCE_TIMEOUT_MS);
        };

        const handleKeyDown = (event: KeyboardEvent) => {
            if (event.defaultPrevented || event.ctrlKey || event.metaKey || event.altKey) return;
            if (isTextEntryTarget(event.target) || isModalOpen()) return;

            if (event.key === HELP_KEY) {
                cancelGoSequence();
                event.preventDefault();
                onShowHelp();
                return;
            }

            if (isAwaitingGoTargetRef.current) {
                const targetRoute = routesByGoSequenceKey[event.key.toLowerCase()];
                cancelGoSequence();
                if (targetRoute) {
                    event.preventDefault();
                    navigate(targetRoute);
                }
                return;
            }

            if (event.key === GO_SEQUENCE_KEY) startGoSequence();
        };

        window.addEventListener('keydown', handleKeyDown);
        return () => {
            cancelGoSequence();
            window.removeEventListener('keydown', handleKeyDown);
        };
    }, [navigate, onShowHelp]);
};
