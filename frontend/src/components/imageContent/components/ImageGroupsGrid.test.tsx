import { act, fireEvent, renderHook, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { buildImageLibraryItem } from '../imageLibraryTestFixtures';
import type { ImageDateGroup } from '../imageDateGroups';
import { useImageSelection, type ImageSelection } from '../useImageSelection';
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

const openButtonAt = (index: number) =>
    screen.getAllByRole('button', { name: 'IMAGES_OPEN_IMAGE_ARIA' })[index]!;

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

    describe('selection', () => {
        const buildSelection = (): { current: ImageSelection } => {
            const { result } = renderHook(() => useImageSelection('scope'));
            return result;
        };

        const renderSelectableGrid = (onOpenImage = jest.fn()) => {
            const selection = renderHook(() => useImageSelection('scope'));
            const view = renderWithoutBackend(
                <ImageGroupsGrid
                    groups={[buildGroup()]}
                    onOpenImage={onOpenImage}
                    onToggleStar={jest.fn()}
                    selection={selection.result.current}
                />
            );
            return { selection, view, onOpenImage };
        };

        it('renders a checkbox per photo and a month checkbox', () => {
            const selection = buildSelection();
            renderWithoutBackend(
                <ImageGroupsGrid
                    groups={[buildGroup()]}
                    onOpenImage={jest.fn()}
                    onToggleStar={jest.fn()}
                    selection={selection.current}
                />
            );

            expect(screen.getAllByRole('checkbox', { name: 'IMAGES_SELECT_IMAGE_ARIA' })[0]!).toBeInTheDocument();
            expect(screen.getByRole('checkbox', { name: 'IMAGES_SELECT_MONTH_ARIA' })).toBeInTheDocument();
        });

        it('opens the viewer on plain click when nothing is selected', () => {
            const { onOpenImage } = renderSelectableGrid();

            fireEvent.click(openButtonAt(0));

            expect(onOpenImage).toHaveBeenCalledWith(1);
        });

        it('toggles instead of opening with ctrl or meta click', () => {
            const toggle = jest.fn();
            const onOpenImage = jest.fn();
            const selection = { ...buildSelection().current, toggle };
            renderWithoutBackend(
                <ImageGroupsGrid
                    groups={[buildGroup()]}
                    onOpenImage={onOpenImage}
                    onToggleStar={jest.fn()}
                    selection={selection}
                />
            );

            fireEvent.click(openButtonAt(0), { ctrlKey: true });
            fireEvent.click(openButtonAt(1), { metaKey: true });

            expect(toggle).toHaveBeenCalledTimes(2);
            expect(onOpenImage).not.toHaveBeenCalled();
        });

        it('toggles on click while a selection exists', () => {
            const toggle = jest.fn();
            const onOpenImage = jest.fn();
            const selection = { ...buildSelection().current, hasSelection: true, toggle };
            renderWithoutBackend(
                <ImageGroupsGrid
                    groups={[buildGroup()]}
                    onOpenImage={onOpenImage}
                    onToggleStar={jest.fn()}
                    selection={selection}
                />
            );

            fireEvent.click(openButtonAt(0));

            expect(toggle).toHaveBeenCalled();
            expect(onOpenImage).not.toHaveBeenCalled();
        });

        it('selects a range with shift click on the tile and on the checkbox', () => {
            const selectRange = jest.fn();
            const selection = { ...buildSelection().current, selectRange };
            renderWithoutBackend(
                <ImageGroupsGrid
                    groups={[buildGroup()]}
                    onOpenImage={jest.fn()}
                    onToggleStar={jest.fn()}
                    selection={selection}
                />
            );

            fireEvent.click(openButtonAt(1), { shiftKey: true });
            fireEvent.click(screen.getAllByRole('checkbox', { name: 'IMAGES_SELECT_IMAGE_ARIA' })[0]!, {
                shiftKey: true,
            });

            expect(selectRange).toHaveBeenCalledTimes(2);
            expect(selectRange.mock.calls[0]?.[1]).toHaveLength(2);
        });

        it('toggles through the photo checkbox and the month checkbox', () => {
            const toggle = jest.fn();
            const toggleMonth = jest.fn();
            const selection = { ...buildSelection().current, toggle, toggleMonth };
            renderWithoutBackend(
                <ImageGroupsGrid
                    groups={[buildGroup()]}
                    onOpenImage={jest.fn()}
                    onToggleStar={jest.fn()}
                    selection={selection}
                />
            );

            fireEvent.click(screen.getAllByRole('checkbox', { name: 'IMAGES_SELECT_IMAGE_ARIA' })[0]!);
            fireEvent.click(screen.getByRole('checkbox', { name: 'IMAGES_SELECT_MONTH_ARIA' }));

            expect(toggle).toHaveBeenCalledTimes(1);
            expect(toggleMonth).toHaveBeenCalledWith(buildGroup().items);
        });

        it('marks the month checkbox as indeterminate when partially selected', () => {
            const selection = {
                ...buildSelection().current,
                readMonthSelectionState: () => 'partial' as const,
            };
            renderWithoutBackend(
                <ImageGroupsGrid
                    groups={[buildGroup()]}
                    onOpenImage={jest.fn()}
                    onToggleStar={jest.fn()}
                    selection={selection}
                />
            );

            const monthCheckbox = screen.getByRole('checkbox', {
                name: 'IMAGES_SELECT_MONTH_ARIA',
            }) as HTMLInputElement;
            expect(monthCheckbox.indeterminate).toBe(true);
        });

        describe('long press', () => {
            const originalPointerEvent = window.PointerEvent;

            beforeEach(() => {
                jest.useFakeTimers();
                class PointerEventWithType extends MouseEvent {
                    readonly pointerType: string;
                    constructor(type: string, init: MouseEventInit & { pointerType?: string } = {}) {
                        super(type, init);
                        this.pointerType = init.pointerType ?? '';
                    }
                }
                window.PointerEvent = PointerEventWithType as unknown as typeof PointerEvent;
            });

            afterEach(() => {
                jest.useRealTimers();
                window.PointerEvent = originalPointerEvent;
            });

            it('starts selection on a touch long press without opening the viewer', () => {
                const toggle = jest.fn();
                const onOpenImage = jest.fn();
                const selection = { ...buildSelection().current, toggle };
                renderWithoutBackend(
                    <ImageGroupsGrid
                        groups={[buildGroup()]}
                        onOpenImage={onOpenImage}
                        onToggleStar={jest.fn()}
                        selection={selection}
                    />
                );
                const photoButton = openButtonAt(0);

                fireEvent.pointerDown(photoButton, { pointerType: 'touch' });
                act(() => {
                    jest.advanceTimersByTime(600);
                });
                fireEvent.pointerUp(photoButton, { pointerType: 'touch' });
                fireEvent.click(photoButton);

                expect(toggle).toHaveBeenCalledTimes(1);
                expect(onOpenImage).not.toHaveBeenCalled();
            });

            it('ignores a short touch and mouse presses', () => {
                const toggle = jest.fn();
                const selection = { ...buildSelection().current, toggle };
                renderWithoutBackend(
                    <ImageGroupsGrid
                        groups={[buildGroup()]}
                        onOpenImage={jest.fn()}
                        onToggleStar={jest.fn()}
                        selection={selection}
                    />
                );
                const photoButton = openButtonAt(0);

                fireEvent.pointerDown(photoButton, { pointerType: 'touch' });
                fireEvent.pointerUp(photoButton, { pointerType: 'touch' });
                fireEvent.pointerDown(photoButton, { pointerType: 'mouse' });
                act(() => {
                    jest.advanceTimersByTime(600);
                });

                expect(toggle).not.toHaveBeenCalled();
            });
        });
    });
});
