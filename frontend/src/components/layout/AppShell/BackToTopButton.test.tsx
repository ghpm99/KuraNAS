import { act, fireEvent, render, screen } from '@testing-library/react';
import { createRef } from 'react';
import { BackToTopButton } from './BackToTopButton';

const buildScrollElement = () => {
    const scrollElement = document.createElement('div');
    Object.defineProperty(scrollElement, 'clientHeight', { configurable: true, value: 500 });
    scrollElement.scrollTo = jest.fn() as unknown as typeof scrollElement.scrollTo;
    return scrollElement;
};

const scrollElementTo = (scrollElement: HTMLElement, scrollTop: number) =>
    act(() => {
        scrollElement.scrollTop = scrollTop;
        scrollElement.dispatchEvent(new Event('scroll'));
    });

describe('layout/AppShell/BackToTopButton', () => {
    it('renders nothing when the scroll element is absent', () => {
        const { container } = render(
            <BackToTopButton scrollElementRef={createRef<HTMLElement>()} isPlayerVisible={false} />
        );

        expect(container).toBeEmptyDOMElement();
    });

    it('stays hidden until scrolling beyond two viewport heights', () => {
        const scrollElement = buildScrollElement();
        render(
            <BackToTopButton scrollElementRef={{ current: scrollElement }} isPlayerVisible={false} />
        );

        scrollElementTo(scrollElement, 1000);
        expect(screen.queryByRole('button')).not.toBeInTheDocument();

        scrollElementTo(scrollElement, 1001);
        expect(screen.getByRole('button', { name: 'BACK_TO_TOP' })).toBeInTheDocument();

        scrollElementTo(scrollElement, 10);
        expect(screen.queryByRole('button')).not.toBeInTheDocument();
    });

    it('scrolls smoothly to the top by default', () => {
        const scrollElement = buildScrollElement();
        render(
            <BackToTopButton scrollElementRef={{ current: scrollElement }} isPlayerVisible={true} />
        );
        scrollElementTo(scrollElement, 2000);

        fireEvent.click(screen.getByRole('button', { name: 'BACK_TO_TOP' }));

        expect(scrollElement.scrollTo).toHaveBeenCalledWith({ top: 0, behavior: 'smooth' });
    });

    it('scrolls instantly when the user prefers reduced motion', () => {
        const originalMatchMedia = window.matchMedia;
        window.matchMedia = ((query: string) => ({
            matches: query.includes('prefers-reduced-motion'),
            media: query,
            addEventListener: jest.fn(),
            removeEventListener: jest.fn(),
            addListener: jest.fn(),
            removeListener: jest.fn(),
        })) as unknown as typeof window.matchMedia;
        const scrollElement = buildScrollElement();
        render(
            <BackToTopButton scrollElementRef={{ current: scrollElement }} isPlayerVisible={false} />
        );
        scrollElementTo(scrollElement, 2000);

        fireEvent.click(screen.getByRole('button', { name: 'BACK_TO_TOP' }));

        expect(scrollElement.scrollTo).toHaveBeenCalledWith({ top: 0, behavior: 'auto' });
        window.matchMedia = originalMatchMedia;
    });
});
