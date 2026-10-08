import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { buildImageLibraryItem } from '../imageLibraryTestFixtures';
import type { ImageDateGroup } from '../imageDateGroups';
import ImageGroupsGrid from './ImageGroupsGrid';

const buildGroup = (overrides: Partial<ImageDateGroup> = {}): ImageDateGroup => ({
    key: '2026-3',
    label: 'March 2026',
    totalCount: null,
    items: [
        buildImageLibraryItem({ file_id: 1, name: 'One.jpg' }),
        buildImageLibraryItem({
            file_id: 2,
            name: 'Two.png',
            starred: true,
            width: 600,
            height: 900,
        }),
    ],
    ...overrides,
});

describe('ImageGroupsGrid', () => {
    it('renders without any backend and with an undefined group list', () => {
        renderWithoutBackend(
            <ImageGroupsGrid
                groups={undefined as unknown as ImageDateGroup[]}
                onOpenImage={jest.fn()}
                onToggleStar={jest.fn()}
            />
        );

        expect(document.querySelectorAll('img')).toHaveLength(0);
    });

    it('renders month headers with the server total and one card per image', () => {
        renderWithoutBackend(
            <ImageGroupsGrid
                groups={[
                    buildGroup({ totalCount: 40 }),
                    buildGroup({ key: 'undated', label: 'Undated' }),
                ]}
                onOpenImage={jest.fn()}
                onToggleStar={jest.fn()}
            />
        );

        expect(screen.getByRole('heading', { name: 'March 2026' })).toBeInTheDocument();
        expect(screen.getByRole('heading', { name: 'Undated' })).toBeInTheDocument();
        expect(screen.getAllByRole('img', { name: 'One.jpg' })).toHaveLength(2);
    });

    it('requests the grid box size with 1x and 2x srcset candidates', () => {
        renderWithoutBackend(
            <ImageGroupsGrid
                groups={[buildGroup()]}
                onOpenImage={jest.fn()}
                onToggleStar={jest.fn()}
            />
        );

        const firstThumbnail = screen.getByRole('img', { name: 'One.jpg' });
        expect(firstThumbnail.getAttribute('src')).toContain(
            '/files/thumbnail/1?width=400&height=400'
        );
        expect(firstThumbnail.getAttribute('srcset')).toContain('width=800&height=800 800w');
    });

    it('renders a headerless group without a heading', () => {
        renderWithoutBackend(
            <ImageGroupsGrid
                groups={[buildGroup({ label: '' })]}
                onOpenImage={jest.fn()}
                onToggleStar={jest.fn()}
            />
        );

        expect(screen.queryByRole('heading')).not.toBeInTheDocument();
    });

    it('opens an image and toggles its star independently', () => {
        const onOpenImage = jest.fn();
        const onToggleStar = jest.fn();
        renderWithoutBackend(
            <ImageGroupsGrid
                groups={[buildGroup()]}
                onOpenImage={onOpenImage}
                onToggleStar={onToggleStar}
            />
        );

        const openButtons = screen.getAllByRole('button', { name: 'IMAGES_OPEN_IMAGE_ARIA' });
        fireEvent.click(openButtons[1]!);
        expect(onOpenImage).toHaveBeenCalledWith(2);

        const addStar = screen.getByRole('button', { name: 'IMAGES_STAR_ADD_ARIA' });
        const removeStar = screen.getByRole('button', { name: 'IMAGES_STAR_REMOVE_ARIA' });
        expect(addStar).toHaveAttribute('aria-pressed', 'false');
        expect(removeStar).toHaveAttribute('aria-pressed', 'true');
        fireEvent.click(addStar);
        fireEvent.click(removeStar);
        expect(onToggleStar).toHaveBeenNthCalledWith(1, 1, false);
        expect(onToggleStar).toHaveBeenNthCalledWith(2, 2, true);
        expect(onOpenImage).toHaveBeenCalledTimes(1);
    });
});
