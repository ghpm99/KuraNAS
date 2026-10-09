import { useEffect, useRef } from 'react';
import {
    isModalOpen,
    isNativelyActivatableTarget,
    isTextEntryTarget,
} from '@/components/shortcuts/keyboardEventContext';
import { useRegisterPageShortcuts } from '@/components/shortcuts/shortcutRegistry';
import {
    GO_SEQUENCE_TIMEOUT_MS,
    routesByGoSequenceKey,
} from '@/components/shortcuts/useGlobalShortcuts';
import { useGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import { nextRepeatMode } from './nextRepeatMode';
import { playerShortcutDefinitions } from './playerShortcutDefinitions';
import { resolvePlayerShortcutAction, type PlayerShortcutAction } from './playerShortcutAction';

const SEEK_STEP_SECONDS = 10;
const VOLUME_STEP = 0.05;
const UNMUTE_VOLUME = 0.7;
const GO_SEQUENCE_KEY = 'g';

const clamp = (value: number, minimum: number, maximum: number): number =>
    Math.min(Math.max(value, minimum), maximum);

export const usePlayerShortcuts = () => {
    const player = useGlobalMusic();
    const playerRef = useRef(player);
    useEffect(() => {
        playerRef.current = player;
    });
    const goSequenceStartedAtRef = useRef<number | null>(null);

    useRegisterPageShortcuts(playerShortcutDefinitions, player.hasQueue);

    useEffect(() => {
        if (!player.hasQueue) return undefined;

        const runAction = (action: PlayerShortcutAction) => {
            const { currentTime, duration, volume, repeatMode } = playerRef.current;
            const currentPosition = Number.isFinite(currentTime) ? currentTime : 0;
            const maximumPosition = Number.isFinite(duration) && duration > 0 ? duration : Infinity;
            const actionsByName: Record<PlayerShortcutAction, () => void> = {
                togglePlayPause: playerRef.current.togglePlayPause,
                next: playerRef.current.next,
                previous: playerRef.current.previous,
                seekForward: () =>
                    playerRef.current.seek(
                        clamp(currentPosition + SEEK_STEP_SECONDS, 0, maximumPosition)
                    ),
                seekBackward: () =>
                    playerRef.current.seek(
                        clamp(currentPosition - SEEK_STEP_SECONDS, 0, maximumPosition)
                    ),
                volumeUp: () => playerRef.current.setVolume(clamp(volume + VOLUME_STEP, 0, 1)),
                volumeDown: () => playerRef.current.setVolume(clamp(volume - VOLUME_STEP, 0, 1)),
                toggleMute: () => playerRef.current.setVolume(volume > 0 ? 0 : UNMUTE_VOLUME),
                toggleShuffle: playerRef.current.toggleShuffle,
                cycleRepeat: () => playerRef.current.setRepeatMode(nextRepeatMode(repeatMode)),
            };
            actionsByName[action]();
        };

        const isCompletingGoSequence = (event: KeyboardEvent): boolean => {
            const startedAt = goSequenceStartedAtRef.current;
            goSequenceStartedAtRef.current = null;
            if (startedAt === null || Date.now() - startedAt > GO_SEQUENCE_TIMEOUT_MS) return false;
            return routesByGoSequenceKey[event.key.toLowerCase()] !== undefined;
        };

        const handleKeyDown = (event: KeyboardEvent) => {
            if (event.defaultPrevented || event.ctrlKey || event.metaKey || event.altKey) return;
            if (isTextEntryTarget(event.target) || isModalOpen()) return;
            if (isCompletingGoSequence(event)) return;
            if (event.key === GO_SEQUENCE_KEY) {
                goSequenceStartedAtRef.current = Date.now();
                return;
            }

            const action = resolvePlayerShortcutAction(event);
            if (!action) return;
            if (action === 'togglePlayPause' && isNativelyActivatableTarget(event.target)) return;
            event.preventDefault();
            runAction(action);
        };

        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [player.hasQueue]);
};
