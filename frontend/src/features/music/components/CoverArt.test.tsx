import { fireEvent, render, screen } from '@testing-library/react';
import CoverArt from './CoverArt';

describe('CoverArt', () => {
    it('renders the cover image without any service mock', () => {
        const { container } = render(
            <CoverArt
                src="/api/v1/music/tracks/1/cover?size=96"
                fallback={<span>placeholder</span>}
            />
        );

        const coverImage = container.querySelector('img');
        expect(coverImage).not.toBeNull();
        expect(coverImage?.getAttribute('src')).toBe('/api/v1/music/tracks/1/cover?size=96');
        expect(screen.queryByText('placeholder')).toBeNull();
    });

    it('shows the placeholder when the cover request fails', () => {
        const { container } = render(
            <CoverArt
                src="/api/v1/music/tracks/1/cover?size=96"
                fallback={<span>placeholder</span>}
            />
        );

        fireEvent.error(container.querySelector('img') as HTMLImageElement);

        expect(screen.getByText('placeholder')).toBeTruthy();
        expect(container.querySelector('img')).toBeNull();
    });

    it('retries the image when the source changes after a failure', () => {
        const { container, rerender } = render(
            <CoverArt src="/cover/1" fallback={<span>placeholder</span>} />
        );
        fireEvent.error(container.querySelector('img') as HTMLImageElement);

        rerender(<CoverArt src="/cover/2" fallback={<span>placeholder</span>} />);

        expect(container.querySelector('img')?.getAttribute('src')).toBe('/cover/2');
    });
});
