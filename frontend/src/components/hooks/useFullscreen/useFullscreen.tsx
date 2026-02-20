import { useState, useEffect, useCallback } from 'react';

interface UseFullscreenReturn {
	isFullscreen: boolean;
	toggleFullscreen: () => void;
	enterFullscreen: () => void;
	exitFullscreen: () => void;
}

export const useFullscreen = (): UseFullscreenReturn => {
	const [isFullscreen, setIsFullscreen] = useState(false);

	const updateFullscreenState = useCallback(() => {
		setIsFullscreen(!!document.fullscreenElement);
	}, []);

	useEffect(() => {
		const handleFullscreenChange = () => {
			updateFullscreenState();
		};

		const handleFullscreenError = (event: Event) => {
			console.error('Fullscreen error:', event);
		};

		document.addEventListener('fullscreenchange', handleFullscreenChange);
		document.addEventListener('fullscreenerror', handleFullscreenError);

		// Initial state check
		updateFullscreenState();

		return () => {
			document.removeEventListener('fullscreenchange', handleFullscreenChange);
			document.removeEventListener('fullscreenerror', handleFullscreenError);
		};
	}, [updateFullscreenState]);

	const enterFullscreen = useCallback(() => {
		const element = document.documentElement;
		
		if (element.requestFullscreen) {
			element.requestFullscreen();
		} else if ((element as any).webkitRequestFullscreen) {
			(element as any).webkitRequestFullscreen();
		} else if ((element as any).mozRequestFullScreen) {
			(element as any).mozRequestFullScreen();
		} else if ((element as any).msRequestFullscreen) {
			(element as any).msRequestFullscreen();
		}
	}, []);

	const exitFullscreen = useCallback(() => {
		if (document.exitFullscreen) {
			document.exitFullscreen();
		} else if ((document as any).webkitExitFullscreen) {
			(document as any).webkitExitFullscreen();
		} else if ((document as any).mozCancelFullScreen) {
			(document as any).mozCancelFullScreen();
		} else if ((document as any).msExitFullscreen) {
			(document as any).msExitFullscreen();
		}
	}, []);

	const toggleFullscreen = useCallback(() => {
		if (isFullscreen) {
			exitFullscreen();
		} else {
			enterFullscreen();
		}
	}, [isFullscreen, enterFullscreen, exitFullscreen]);

	return {
		isFullscreen,
		toggleFullscreen,
		enterFullscreen,
		exitFullscreen,
	};
};
