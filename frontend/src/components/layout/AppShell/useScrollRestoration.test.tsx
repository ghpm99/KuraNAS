import { act, fireEvent, render } from '@testing-library/react';
import { useRef } from 'react';
import { MemoryRouter, useNavigate } from 'react-router-dom';
import { clearSavedScrollPositions, useScrollRestoration } from './useScrollRestoration';

let navigateTo: ReturnType<typeof useNavigate>;

const Harness = () => {
    const scrollAreaRef = useRef<HTMLDivElement>(null);
    navigateTo = useNavigate();
    useScrollRestoration(scrollAreaRef);
    return <div ref={scrollAreaRef} data-testid="scroll-area" />;
};

const renderHarness = () => {
    const rendered = render(
        <MemoryRouter initialEntries={['/first']}>
            <Harness />
        </MemoryRouter>
    );
    return rendered.getByTestId('scroll-area');
};

const defineContentHeight = (scrollArea: HTMLElement, scrollHeight: number, clientHeight = 500) => {
    Object.defineProperty(scrollArea, 'scrollHeight', { configurable: true, value: scrollHeight });
    Object.defineProperty(scrollArea, 'clientHeight', { configurable: true, value: clientHeight });
};

const trackScrollTopWrites = (scrollArea: HTMLElement) => {
    const writes: number[] = [];
    let currentScrollTop = scrollArea.scrollTop;
    Object.defineProperty(scrollArea, 'scrollTop', {
        configurable: true,
        get: () => currentScrollTop,
        set: (nextScrollTop: number) => {
            writes.push(nextScrollTop);
            currentScrollTop = nextScrollTop;
        },
    });
    return writes;
};

const scrollTo = (scrollArea: HTMLElement, scrollTop: number) => {
    scrollArea.scrollTop = scrollTop;
    fireEvent.scroll(scrollArea);
};

describe('layout/AppShell/useScrollRestoration', () => {
    beforeEach(() => {
        clearSavedScrollPositions();
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    it('mounts without a scroll element value being set', () => {
        const scrollArea = renderHarness();

        expect(scrollArea.scrollTop).toBe(0);
    });

    it('resets scroll to the top on a pushed navigation', () => {
        const scrollArea = renderHarness();
        defineContentHeight(scrollArea, 5000);
        scrollTo(scrollArea, 900);

        act(() => navigateTo('/second'));

        expect(scrollArea.scrollTop).toBe(0);
    });

    it('keeps the scroll position when a replace does not change the pathname', () => {
        const scrollArea = renderHarness();
        defineContentHeight(scrollArea, 5000);
        scrollTo(scrollArea, 900);

        act(() => navigateTo('/first?page=2', { replace: true }));

        expect(scrollArea.scrollTop).toBe(900);
    });

    it('resets scroll when a replace changes the pathname', () => {
        const scrollArea = renderHarness();
        defineContentHeight(scrollArea, 5000);
        scrollTo(scrollArea, 900);

        act(() => navigateTo('/other', { replace: true }));

        expect(scrollArea.scrollTop).toBe(0);
    });

    it('restores the saved position when going back', () => {
        const scrollArea = renderHarness();
        defineContentHeight(scrollArea, 5000);
        act(() => navigateTo('/second'));
        scrollTo(scrollArea, 700);

        act(() => navigateTo('/third'));
        expect(scrollArea.scrollTop).toBe(0);

        act(() => navigateTo(-1));
        expect(scrollArea.scrollTop).toBe(700);
    });

    it('retries across frames until the content is tall enough', () => {
        jest.useFakeTimers();
        const scrollArea = renderHarness();
        defineContentHeight(scrollArea, 5000);
        act(() => navigateTo('/second'));
        scrollTo(scrollArea, 800);
        act(() => navigateTo('/third'));
        const scrollTopWrites = trackScrollTopWrites(scrollArea);

        defineContentHeight(scrollArea, 600);
        act(() => navigateTo(-1));
        act(() => {
            jest.advanceTimersByTime(64);
        });
        const writesWhileContentIsShort = scrollTopWrites.length;
        expect(writesWhileContentIsShort).toBeGreaterThan(1);

        defineContentHeight(scrollArea, 5000);
        act(() => {
            jest.advanceTimersByTime(64);
        });

        expect(scrollTopWrites.length).toBeGreaterThan(writesWhileContentIsShort);
        expect(scrollArea.scrollTop).toBe(800);

        const writesAfterRestore = scrollTopWrites.length;
        act(() => {
            jest.advanceTimersByTime(500);
        });
        expect(scrollTopWrites.length).toBe(writesAfterRestore);
    });

    it('gives up after the frame budget when the content never grows', () => {
        jest.useFakeTimers();
        const scrollArea = renderHarness();
        defineContentHeight(scrollArea, 5000);
        act(() => navigateTo('/second'));
        scrollTo(scrollArea, 800);
        act(() => navigateTo('/third'));
        const scrollTopWrites = trackScrollTopWrites(scrollArea);

        defineContentHeight(scrollArea, 600);
        act(() => navigateTo(-1));
        act(() => {
            jest.advanceTimersByTime(2000);
        });

        expect(scrollTopWrites.length).toBe(30);
    });
});
