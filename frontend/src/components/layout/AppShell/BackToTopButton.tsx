import { useEffect, useState, type RefObject } from 'react';
import { IconButton, useMediaQuery } from '@mui/material';
import { ArrowUp } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import styles from './BackToTopButton.module.css';

const VISIBLE_AFTER_VIEWPORT_HEIGHTS = 2;

interface BackToTopButtonProps {
    scrollElementRef: RefObject<HTMLElement | null>;
    isPlayerVisible: boolean;
}

export const BackToTopButton = ({ scrollElementRef, isPlayerVisible }: BackToTopButtonProps) => {
    const { t } = useI18n();
    const [isVisible, setIsVisible] = useState(false);
    const prefersReducedMotion = useMediaQuery('(prefers-reduced-motion: reduce)');

    useEffect(() => {
        const scrollElement = scrollElementRef.current;
        if (!scrollElement) return;

        const updateVisibility = () =>
            setIsVisible(
                scrollElement.scrollTop >
                    scrollElement.clientHeight * VISIBLE_AFTER_VIEWPORT_HEIGHTS
            );

        scrollElement.addEventListener('scroll', updateVisibility, { passive: true });
        return () => scrollElement.removeEventListener('scroll', updateVisibility);
    }, [scrollElementRef]);

    if (!isVisible) return null;

    const handleScrollToTop = () =>
        scrollElementRef.current?.scrollTo({
            top: 0,
            behavior: prefersReducedMotion ? 'auto' : 'smooth',
        });

    const className = isPlayerVisible
        ? `${styles.button} ${styles.buttonWithPlayer}`
        : styles.button;

    return (
        <IconButton
            className={className}
            onClick={handleScrollToTop}
            title={t('BACK_TO_TOP')}
            aria-label={t('BACK_TO_TOP')}
        >
            <ArrowUp size={20} />
        </IconButton>
    );
};
