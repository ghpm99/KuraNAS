import { useRef, useState, useCallback } from 'react';

interface UseIntersectionObserverOptions {
	threshold?: number;
	rootMargin?: string;
	enabled?: boolean;
	onIntersect?: () => void;
}

export function useIntersectionObserver<T extends HTMLElement = HTMLDivElement>({
	threshold = 0.1,
	rootMargin = '100px',
	enabled = true,
	onIntersect,
}: UseIntersectionObserverOptions = {}) {
	const [isIntersecting, setIsIntersecting] = useState(false);
	const observerRef = useRef<IntersectionObserver | null>(null);

	const ref = useCallback(
		(node: T | null) => {
			if (!enabled || !node) {
				if (observerRef.current) {
					observerRef.current.disconnect();
					observerRef.current = null;
				}
				return;
			}

			const observer = new IntersectionObserver(
				([entry]) => {
					const intersecting = entry?.isIntersecting ?? false;
					setIsIntersecting(intersecting);
					if (intersecting && onIntersect) {
						onIntersect();
					}
				},
				{ threshold, rootMargin },
			);

			observerRef.current = observer;
			observer.observe(node);

			return () => {
				observer.disconnect();
				observerRef.current = null;
			};
		},
		[enabled, threshold, rootMargin, onIntersect],
	);

	return { ref, isIntersecting };
}
