import { fireEvent, screen, within } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import type { ImageTimelineBucket } from '@/types/imageLibrary';
import ImageDateScrubber from './ImageDateScrubber';

const buckets: ImageTimelineBucket[] = [
    { year: 2026, month: 3, count: 5 },
    { year: 2026, month: 1, count: 2 },
    { year: 2024, month: 12, count: 9 },
];

const mockMatchMedia = (isDesktop: boolean) => {
    Object.defineProperty(window, 'matchMedia', {
        configurable: true,
        writable: true,
        value: (query: string) => ({
            matches: isDesktop,
            media: query,
            addEventListener: jest.fn(),
            removeEventListener: jest.fn(),
            addListener: jest.fn(),
            removeListener: jest.fn(),
            dispatchEvent: jest.fn(),
            onchange: null,
        }),
    });
};

describe('ImageDateScrubber', () => {
    afterEach(() => {
        Reflect.deleteProperty(window, 'matchMedia');
    });

    it('renders nothing without any backend or timeline', () => {
        const { container } = renderWithoutBackend(
            <ImageDateScrubber
                buckets={undefined as unknown as ImageTimelineBucket[]}
                onSelectMonth={jest.fn()}
            />
        );

        expect(container).toBeEmptyDOMElement();
    });

    it('on phones opens a month sheet and jumps to the chosen month', () => {
        mockMatchMedia(false);
        const onSelectMonth = jest.fn();
        renderWithoutBackend(<ImageDateScrubber buckets={buckets} onSelectMonth={onSelectMonth} />);

        fireEvent.click(screen.getByRole('button', { name: 'IMAGES_SCRUBBER_OPEN' }));
        const sheet = screen.getByRole('presentation');
        expect(within(sheet).getByRole('heading', { name: '2026' })).toBeInTheDocument();
        expect(within(sheet).getByRole('heading', { name: '2024' })).toBeInTheDocument();

        const monthButtons = within(sheet).getAllByRole('button', {
            name: 'IMAGES_SCRUBBER_MONTH_ARIA',
        });
        expect(monthButtons).toHaveLength(3);
        fireEvent.click(monthButtons[0]!);

        expect(onSelectMonth).toHaveBeenCalledWith(2026, 3);
    });

    it('on desktop shows a year rail whose months open in a panel', () => {
        mockMatchMedia(true);
        const onSelectMonth = jest.fn();
        renderWithoutBackend(<ImageDateScrubber buckets={buckets} onSelectMonth={onSelectMonth} />);

        const rail = screen.getByRole('navigation', { name: 'IMAGES_SCRUBBER_ARIA' });
        const yearButtons = within(rail).getAllByRole('button');
        expect(yearButtons.map((button) => button.textContent)).toEqual(['2026', '2024']);
        expect(
            within(rail).queryByRole('button', { name: 'IMAGES_SCRUBBER_MONTH_ARIA' })
        ).toBeNull();

        fireEvent.click(yearButtons[1]!);
        expect(yearButtons[1]).toHaveAttribute('aria-expanded', 'true');
        fireEvent.click(within(rail).getByRole('button', { name: 'IMAGES_SCRUBBER_MONTH_ARIA' }));

        expect(onSelectMonth).toHaveBeenCalledWith(2024, 12);
        expect(screen.queryByRole('button', { name: 'IMAGES_SCRUBBER_MONTH_ARIA' })).toBeNull();
    });

    it('collapses the expanded year when it is clicked again', () => {
        mockMatchMedia(true);
        renderWithoutBackend(<ImageDateScrubber buckets={buckets} onSelectMonth={jest.fn()} />);

        const yearButton = screen.getByRole('button', { name: '2026' });
        fireEvent.click(yearButton);
        expect(screen.getAllByRole('button', { name: 'IMAGES_SCRUBBER_MONTH_ARIA' })).toHaveLength(
            2
        );

        fireEvent.click(yearButton);
        expect(yearButton).toHaveAttribute('aria-expanded', 'false');
    });
});
