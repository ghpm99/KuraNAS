import { useVideoPlayer } from '@/components/hooks/videoPlayerProvider/videoPlayerProvider';
import { useEffect, useCallback } from 'react';

interface UseKeyboardShortcutsReturn {
	// No return value needed, just handles keyboard events
}

export const useKeyboardShortcuts = (): UseKeyboardShortcutsReturn => {
	const {
		isPlaying,
		currentTime,
		duration,
		volume,
		pause,
		resume,
		seekTo,
		setVolume,
		togglePlayPause,
		nextVideo,
		previousVideo,
		toggleFullscreen,
		setPlaybackRate,
	} = useVideoPlayer();

	const handleKeyDown = useCallback((event: KeyboardEvent) => {
		// Ignore if user is typing in an input field
		const target = event.target as HTMLElement;
		if (
			target.tagName === 'INPUT' ||
			target.tagName === 'TEXTAREA' ||
			target.contentEditable === 'true'
		) {
			return;
		}

		// Prevent default for video-related keys
		const videoKeys = [
			' ', // Space
			'ArrowLeft',
			'ArrowRight',
			'ArrowUp',
			'ArrowDown',
			'f',
			'F',
			'm',
			'M',
			'j',
			'J',
			'l',
			'L',
			'k',
			'K',
			'n',
			'N',
			'p',
			'P',
			'0',
			'1',
			'2',
			'3',
			'4',
			'5',
			'6',
			'7',
			'8',
			'9',
			'+',
			'=',
			'-',
			'_',
		];

		if (videoKeys.includes(event.key)) {
			event.preventDefault();
		}

		switch (event.key) {
			// Play/Pause
			case ' ':
			case 'k':
			case 'K':
				togglePlayPause();
				break;

			// Seek backward/forward (10 seconds)
			case 'ArrowLeft':
			case 'j':
			case 'J':
				seekTo(Math.max(0, currentTime - 10));
				break;
			case 'ArrowRight':
			case 'l':
			case 'L':
				seekTo(Math.min(duration, currentTime + 10));
				break;

			// Seek backward/forward (5 seconds - fine control)
			case 'ArrowDown':
				seekTo(Math.max(0, currentTime - 5));
				break;
			case 'ArrowUp':
				seekTo(Math.min(duration, currentTime + 5));
				break;

			// Volume control
			case 'ArrowUp':
				if (event.shiftKey) {
					setVolume(Math.min(1, volume + 0.1));
				}
				break;
			case 'ArrowDown':
				if (event.shiftKey) {
					setVolume(Math.max(0, volume - 0.1));
				}
				break;

			// Mute/Unmute
			case 'm':
			case 'M':
				setVolume(volume === 0 ? 0.5 : 0);
				break;

			// Fullscreen
			case 'f':
			case 'F':
				toggleFullscreen();
				break;

			// Next/Previous video
			case 'n':
			case 'N':
				nextVideo();
				break;
			case 'p':
			case 'P':
				previousVideo();
				break;

			// Seek to percentage (0-9)
			case '0':
				seekTo(0);
				break;
			case '1':
				seekTo(duration * 0.1);
				break;
			case '2':
				seekTo(duration * 0.2);
				break;
			case '3':
				seekTo(duration * 0.3);
				break;
			case '4':
				seekTo(duration * 0.4);
				break;
			case '5':
				seekTo(duration * 0.5);
				break;
			case '6':
				seekTo(duration * 0.6);
				break;
			case '7':
				seekTo(duration * 0.7);
				break;
			case '8':
				seekTo(duration * 0.8);
				break;
			case '9':
				seekTo(duration * 0.9);
				break;

			// Playback speed
			case '+':
			case '=':
				setPlaybackRate(Math.min(2, (volume === 0 ? 1 : volume) + 0.25));
				break;
			case '-':
			case '_':
				setPlaybackRate(Math.max(0.25, (volume === 0 ? 1 : volume) - 0.25));
				break;

			// Home/End keys
			case 'Home':
				seekTo(0);
				break;
			case 'End':
				seekTo(duration);
				break;

			default:
				break;
		}
	}, [
		isPlaying,
		currentTime,
		duration,
		volume,
		pause,
		resume,
		seekTo,
		setVolume,
		togglePlayPause,
		nextVideo,
		previousVideo,
		toggleFullscreen,
		setPlaybackRate,
	]);

	useEffect(() => {
		document.addEventListener('keydown', handleKeyDown);
		return () => {
			document.removeEventListener('keydown', handleKeyDown);
		};
	}, [handleKeyDown]);

	return {};
};
