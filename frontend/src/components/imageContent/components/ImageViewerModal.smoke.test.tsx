import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import type { ImageLibraryItem } from '@/types/imageLibrary';
import { buildImageLibraryItem } from '../imageLibraryTestFixtures';
import ImageViewerModal from './ImageViewerModal';

describe('ImageViewerModal (no-mock render)', () => {
    it('renders without any backend and falls back to unavailable details', () => {
        renderWithoutBackend(
            <ImageViewerModal
                activeImage={buildImageLibraryItem({
                    category: 'unexpected' as ImageLibraryItem['category'],
                })}
                activeIndex={0}
                activeImageDate={null}
                dateFormatter={new Intl.DateTimeFormat('en-US')}
                filteredImages={[buildImageLibraryItem()]}
                zoom={1}
                showDetails
                showFilmstrip={false}
                isSlideshowPlaying={false}
                isFavoritePending={false}
                onToggleDetails={jest.fn()}
                onToggleFilmstrip={jest.fn()}
                onToggleSlideshow={jest.fn()}
                onToggleFavorite={jest.fn()}
                onOpenFolder={jest.fn()}
                onDecreaseZoom={jest.fn()}
                onResetZoom={jest.fn()}
                onIncreaseZoom={jest.fn()}
                onClose={jest.fn()}
                onPrevious={jest.fn()}
                onNext={jest.fn()}
                onOpenImage={jest.fn()}
            />
        );

        expect(screen.getByRole('dialog', { name: 'Trip.jpg' })).toBeInTheDocument();
    });
});
