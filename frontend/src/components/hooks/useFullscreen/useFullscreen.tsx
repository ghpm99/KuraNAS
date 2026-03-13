import { useState, useEffect, useCallback } from 'react';

interface UseFullscreenReturn {
	isFullscreen: boolean;
	toggleFullscreen: () => void;
	enterFullscreen: () => void;
	exitFullscreen: () => void;
}

interface FullscreenElement extends HTMLElement {
	webkitRequestFullscreen?: () => Promise<void> | void;
	mozRequestFullScreen?: () => Promise<void> | void;
	msRequestFullscreen?: () => Promise<void> | void;
}

interface FullscreenDocument extends Document {
	webkitExitFullscreen?: () => Promise<void> | void;
	mozCancelFullScreen?: () => Promise<void> | void;
	msExitFullscreen?: () => Promise<void> | void;
}

export const useFullscreen = (): UseFullscreenReturn => {
	const [isFullscreen, setIsFullscreen] = useState(() => !!document.fullscreenElement);

	useEffect(() => {
		const handleFullscreenChange = () => {
			setIsFullscreen(!!document.fullscreenElement);
		};

		const handleFullscreenError = (event: Event) => {
			console.error('Fullscreen error', event);
		};

		document.addEventListener('fullscreenchange', handleFullscreenChange);
		document.addEventListener('fullscreenerror', handleFullscreenError);

		return () => {
			document.removeEventListener('fullscreenchange', handleFullscreenChange);
			document.removeEventListener('fullscreenerror', handleFullscreenError);
		};
	}, []);

	const enterFullscreen = useCallback(() => {
		const element = document.documentElement as FullscreenElement;
		
		if (element.requestFullscreen) {
			element.requestFullscreen();
		} else if (element.webkitRequestFullscreen) {
			element.webkitRequestFullscreen();
		} else if (element.mozRequestFullScreen) {
			element.mozRequestFullScreen();
		} else if (element.msRequestFullscreen) {
			element.msRequestFullscreen();
		}
	}, []);

	const exitFullscreen = useCallback(() => {
		const fullscreenDocument = document as FullscreenDocument;
		if (document.exitFullscreen) {
			document.exitFullscreen();
		} else if (fullscreenDocument.webkitExitFullscreen) {
			fullscreenDocument.webkitExitFullscreen();
		} else if (fullscreenDocument.mozCancelFullScreen) {
			fullscreenDocument.mozCancelFullScreen();
		} else if (fullscreenDocument.msExitFullscreen) {
			fullscreenDocument.msExitFullscreen();
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
