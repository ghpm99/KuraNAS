import { DEFAULT_UNMUTE_VOLUME } from '@/features/videos/hooks/useVideoPlayer/videoPlaybackPreferences';
import {
    MAX_PLAYBACK_RATE,
    MIN_PLAYBACK_RATE,
    type PlayerShortcutAction,
} from './playerShortcutMapping';

export interface PlayerShortcutFeedback {
    messageKey: string;
    params?: Record<string, string>;
}

export interface PlayerShortcutSnapshot {
    isPlaying: boolean;
    currentTime: number;
    duration: number;
    volume: number;
    isMuted: boolean;
    playbackRate: number;
    canGoNext: boolean;
    canGoPrevious: boolean;
}

export interface PlayerShortcutControls {
    seekTo: (time: number) => void;
    setVolume: (volume: number) => void;
    toggleMute: () => void;
    setPlaybackRate: (rate: number) => void;
    toggleFullscreen: () => void;
    togglePictureInPicture: () => void | Promise<void>;
    togglePlayPause: () => void;
    nextVideo: () => void | Promise<void>;
    previousVideo: () => void | Promise<void>;
    openHelp: () => void;
}

const clamp = (value: number, min: number, max: number) => Math.max(min, Math.min(max, value));

const hasKnownDuration = (duration: number) => Number.isFinite(duration) && duration > 0;

export function clampSeekTime(targetTime: number, duration: number) {
    const nonNegativeTime = Math.max(0, targetTime);
    return hasKnownDuration(duration) ? Math.min(nonNegativeTime, duration) : nonNegativeTime;
}

const seekRelative = (
    seconds: number,
    snapshot: PlayerShortcutSnapshot,
    controls: PlayerShortcutControls
): PlayerShortcutFeedback => {
    controls.seekTo(clampSeekTime(snapshot.currentTime + seconds, snapshot.duration));
    return {
        messageKey:
            seconds > 0
                ? 'VIDEO_SHORTCUT_FEEDBACK_SEEK_FORWARD'
                : 'VIDEO_SHORTCUT_FEEDBACK_SEEK_BACKWARD',
        params: { seconds: String(Math.abs(seconds)) },
    };
};

const seekToFraction = (
    fraction: number,
    snapshot: PlayerShortcutSnapshot,
    controls: PlayerShortcutControls
): PlayerShortcutFeedback | null => {
    if (!hasKnownDuration(snapshot.duration)) return null;
    controls.seekTo(clampSeekTime(snapshot.duration * fraction, snapshot.duration));
    return {
        messageKey: 'VIDEO_SHORTCUT_FEEDBACK_SEEK_PERCENT',
        params: { percent: String(Math.round(fraction * 100)) },
    };
};

const adjustVolume = (
    delta: number,
    snapshot: PlayerShortcutSnapshot,
    controls: PlayerShortcutControls
): PlayerShortcutFeedback => {
    const adjustedVolume = clamp(Math.round((snapshot.volume + delta) * 100) / 100, 0, 1);
    controls.setVolume(adjustedVolume);
    return {
        messageKey: 'VIDEO_SHORTCUT_FEEDBACK_VOLUME',
        params: { percent: String(Math.round(adjustedVolume * 100)) },
    };
};

const toggleMute = (
    snapshot: PlayerShortcutSnapshot,
    controls: PlayerShortcutControls
): PlayerShortcutFeedback => {
    controls.toggleMute();
    const isSilent = snapshot.isMuted || snapshot.volume === 0;
    if (!isSilent) return { messageKey: 'VIDEO_SHORTCUT_FEEDBACK_MUTED' };
    const restoredVolume = snapshot.volume || DEFAULT_UNMUTE_VOLUME;
    return {
        messageKey: 'VIDEO_SHORTCUT_FEEDBACK_VOLUME',
        params: { percent: String(Math.round(restoredVolume * 100)) },
    };
};

const adjustPlaybackRate = (
    delta: number,
    snapshot: PlayerShortcutSnapshot,
    controls: PlayerShortcutControls
): PlayerShortcutFeedback => {
    const adjustedRate = clamp(snapshot.playbackRate + delta, MIN_PLAYBACK_RATE, MAX_PLAYBACK_RATE);
    controls.setPlaybackRate(adjustedRate);
    return {
        messageKey: 'VIDEO_SHORTCUT_FEEDBACK_SPEED',
        params: { rate: String(adjustedRate) },
    };
};

export function runPlayerShortcut(
    action: PlayerShortcutAction,
    snapshot: PlayerShortcutSnapshot,
    controls: PlayerShortcutControls
): PlayerShortcutFeedback | null {
    switch (action.type) {
        case 'togglePlayPause':
            controls.togglePlayPause();
            return {
                messageKey: snapshot.isPlaying
                    ? 'VIDEO_SHORTCUT_FEEDBACK_PAUSE'
                    : 'VIDEO_SHORTCUT_FEEDBACK_PLAY',
            };
        case 'seekRelative':
            return seekRelative(action.seconds, snapshot, controls);
        case 'seekToFraction':
            return seekToFraction(action.fraction, snapshot, controls);
        case 'adjustVolume':
            return adjustVolume(action.delta, snapshot, controls);
        case 'toggleMute':
            return toggleMute(snapshot, controls);
        case 'adjustPlaybackRate':
            return adjustPlaybackRate(action.delta, snapshot, controls);
        case 'toggleFullscreen':
            controls.toggleFullscreen();
            return null;
        case 'togglePictureInPicture':
            void controls.togglePictureInPicture();
            return null;
        case 'nextVideo':
            if (!snapshot.canGoNext) return null;
            void controls.nextVideo();
            return { messageKey: 'VIDEO_SHORTCUT_FEEDBACK_NEXT' };
        case 'previousVideo':
            if (!snapshot.canGoPrevious) return null;
            void controls.previousVideo();
            return { messageKey: 'VIDEO_SHORTCUT_FEEDBACK_PREVIOUS' };
        case 'openHelp':
            controls.openHelp();
            return null;
    }
}
