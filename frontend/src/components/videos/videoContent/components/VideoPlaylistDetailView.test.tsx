import { fireEvent, render, screen } from '@testing-library/react';
import VideoPlaylistDetailView from './VideoPlaylistDetailView';
import type { VideoPlaylistDto, VideoPlaylistItemDto } from '@/service/videoPlayback';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string>) => {
            const map: Record<string, string> = {
                VIDEO_BACK_TO_VIDEOS: 'Back to videos',
                VIDEO_PLAYLIST_META: `${params?.count ?? '0'} items`,
                VIDEO_EDIT_PLAYLIST: 'Edit Playlist',
                VIDEO_DISPLAY_NAME_PLACEHOLDER: 'Display name',
                VIDEO_SAVE_NAME: 'Save Name',
                VIDEO_SOURCE_MANUAL: 'Manual',
                VIDEO_SOURCE_AUTO: 'Auto',
            };
            return map[key] ?? key;
        },
    }),
}));

jest.mock('@/service/apiUrl', () => ({
    getApiV1BaseUrl: () => 'http://localhost:8000/api/v1',
}));

const createPlaylistItem = (
    id: number,
    videoId: number,
    orderIndex: number,
    overrides?: Partial<VideoPlaylistItemDto>
): VideoPlaylistItemDto => ({
    id,
    order_index: orderIndex,
    source_kind: 'auto',
    video: {
        id: videoId,
        name: `Video ${videoId}`,
        path: `/videos/video-${videoId}.mp4`,
        parent_path: '/videos',
        format: 'mp4',
        size: 100000,
    },
    status: 'not_started',
    progress_pct: 0,
    ...overrides,
});

const createPlaylist = (overrides: Partial<VideoPlaylistDto> = {}): VideoPlaylistDto => ({
    id: 1,
    type: 'custom',
    source_path: '/videos',
    name: 'Test Playlist',
    is_hidden: false,
    is_auto: false,
    group_mode: 'single',
    classification: 'series',
    item_count: 3,
    cover_video_id: 10,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    last_played_at: null,
    items: [
        createPlaylistItem(1, 10, 0),
        createPlaylistItem(2, 20, 1),
        createPlaylistItem(3, 30, 2),
    ],
    ...overrides,
});

const defaultProps = () => ({
    playlist: createPlaylist(),
    isRenaming: false,
    isRemoving: false,
    isReordering: false,
    onBack: jest.fn(),
    onOpenVideo: jest.fn(),
    onRename: jest.fn(),
    onRemoveVideo: jest.fn(),
    onMoveItem: jest.fn(),
});

describe('VideoPlaylistDetailView', () => {
    it('renders the playlist name and metadata', () => {
        render(<VideoPlaylistDetailView {...defaultProps()} />);

        expect(screen.getByText('Test Playlist')).toBeInTheDocument();
        expect(screen.getByText('3 items')).toBeInTheDocument();
        expect(screen.getByText('SERIES')).toBeInTheDocument();
    });

    it('renders the hero cover image from cover_video_id', () => {
        render(<VideoPlaylistDetailView {...defaultProps()} />);

        const heroImg = screen.getAllByRole('img')[0];
        expect(heroImg).toHaveAttribute(
            'src',
            'http://localhost:8000/api/v1/files/video-thumbnail/10?width=1280&height=720'
        );
    });

    it('uses first item video id as cover when cover_video_id is null', () => {
        const playlist = createPlaylist({ cover_video_id: null });
        render(<VideoPlaylistDetailView {...defaultProps()} playlist={playlist} />);

        const heroImg = screen.getAllByRole('img')[0];
        expect(heroImg).toHaveAttribute(
            'src',
            'http://localhost:8000/api/v1/files/video-thumbnail/10?width=1280&height=720'
        );
    });

    it('renders all video items in order', () => {
        render(<VideoPlaylistDetailView {...defaultProps()} />);

        expect(screen.getByText('Video 10')).toBeInTheDocument();
        expect(screen.getByText('Video 20')).toBeInTheDocument();
        expect(screen.getByText('Video 30')).toBeInTheDocument();
    });

    it('sorts items by order_index', () => {
        const playlist = createPlaylist({
            items: [
                createPlaylistItem(3, 30, 2),
                createPlaylistItem(1, 10, 0),
                createPlaylistItem(2, 20, 1),
            ],
        });
        const { container } = render(
            <VideoPlaylistDetailView {...defaultProps()} playlist={playlist} />
        );

        const titles = container.querySelectorAll('[class*="detailTitle"]');
        expect(titles[0]).toHaveTextContent('Video 10');
        expect(titles[1]).toHaveTextContent('Video 20');
        expect(titles[2]).toHaveTextContent('Video 30');
    });

    it('calls onBack when back button is clicked', () => {
        const props = defaultProps();
        render(<VideoPlaylistDetailView {...props} />);

        fireEvent.click(screen.getByText('Back to videos'));
        expect(props.onBack).toHaveBeenCalledTimes(1);
    });

    it('calls onOpenVideo with video id when a video item is clicked', () => {
        const props = defaultProps();
        render(<VideoPlaylistDetailView {...props} />);

        // Click the main button for the first video
        const mainButtons = screen.getAllByText('Video 10');
        fireEvent.click(mainButtons[0]);
        expect(props.onOpenVideo).toHaveBeenCalledWith(10);
    });

    it('calls onRename with the name draft when save is clicked', () => {
        const props = defaultProps();
        render(<VideoPlaylistDetailView {...props} />);

        const input = screen.getByDisplayValue('Test Playlist');
        fireEvent.change(input, { target: { value: 'New Name' } });
        fireEvent.click(screen.getByText('Save Name'));

        expect(props.onRename).toHaveBeenCalledWith('New Name');
    });

    it('disables save button when name draft is empty', () => {
        const props = defaultProps();
        render(<VideoPlaylistDetailView {...props} />);

        const input = screen.getByDisplayValue('Test Playlist');
        fireEvent.change(input, { target: { value: '' } });

        const saveBtn = screen.getByText('Save Name');
        expect(saveBtn).toBeDisabled();
    });

    it('disables save button when name draft is only whitespace', () => {
        const props = defaultProps();
        render(<VideoPlaylistDetailView {...props} />);

        const input = screen.getByDisplayValue('Test Playlist');
        fireEvent.change(input, { target: { value: '   ' } });

        const saveBtn = screen.getByText('Save Name');
        expect(saveBtn).toBeDisabled();
    });

    it('disables save button when isRenaming is true', () => {
        const props = { ...defaultProps(), isRenaming: true };
        render(<VideoPlaylistDetailView {...props} />);

        const saveBtn = screen.getByText('Save Name');
        expect(saveBtn).toBeDisabled();
    });

    it('calls onRemoveVideo with correct video id', () => {
        const props = defaultProps();
        render(<VideoPlaylistDetailView {...props} />);

        // There should be 3 delete buttons (one per video item)
        const deleteButtons = screen.getAllByRole('button').filter((btn) => {
            return btn.className.includes('iconBtnDanger');
        });
        expect(deleteButtons).toHaveLength(3);
        fireEvent.click(deleteButtons[0]);
        expect(props.onRemoveVideo).toHaveBeenCalledWith(10);
    });

    it('disables remove buttons when isRemoving is true', () => {
        const props = { ...defaultProps(), isRemoving: true };
        render(<VideoPlaylistDetailView {...props} />);

        const deleteButtons = screen
            .getAllByRole('button')
            .filter((btn) => btn.className.includes('iconBtnDanger'));
        deleteButtons.forEach((btn) => expect(btn).toBeDisabled());
    });

    it('calls onMoveItem with index and direction -1 for move up', () => {
        const props = defaultProps();
        render(<VideoPlaylistDetailView {...props} />);

        // Up buttons: first is disabled (index 0), second should work
        const upButtons = screen
            .getAllByRole('button')
            .filter(
                (btn) => btn.className.includes('iconBtn') && !btn.className.includes('Danger')
            );
        // Up buttons are at even indices (0, 2, 4), down at odd (1, 3, 5)
        // Click the second up button (index 1 item)
        fireEvent.click(upButtons[2]); // third iconBtn is the second "up" button
        expect(props.onMoveItem).toHaveBeenCalledWith(1, -1);
    });

    it('calls onMoveItem with index and direction 1 for move down', () => {
        const props = defaultProps();
        render(<VideoPlaylistDetailView {...props} />);

        // Find down buttons - they have ArrowDown icon
        const actionButtons = screen
            .getAllByRole('button')
            .filter(
                (btn) => btn.className.includes('iconBtn') && !btn.className.includes('Danger')
            );
        // First down button is actionButtons[1]
        fireEvent.click(actionButtons[1]);
        expect(props.onMoveItem).toHaveBeenCalledWith(0, 1);
    });

    it('disables first item up button', () => {
        const props = defaultProps();
        render(<VideoPlaylistDetailView {...props} />);

        const upButtons = screen
            .getAllByRole('button')
            .filter(
                (btn) => btn.className.includes('iconBtn') && !btn.className.includes('Danger')
            );
        // First up button (index 0) should be disabled
        expect(upButtons[0]).toBeDisabled();
    });

    it('disables last item down button', () => {
        const props = defaultProps();
        render(<VideoPlaylistDetailView {...props} />);

        const downButtons = screen
            .getAllByRole('button')
            .filter(
                (btn) => btn.className.includes('iconBtn') && !btn.className.includes('Danger')
            );
        // Last down button should be disabled (last item, index 2)
        // Buttons pattern per item: up, down. So index 5 is last down
        expect(downButtons[5]).toBeDisabled();
    });

    it('disables move buttons when isReordering is true', () => {
        const props = { ...defaultProps(), isReordering: true };
        render(<VideoPlaylistDetailView {...props} />);

        const moveButtons = screen
            .getAllByRole('button')
            .filter(
                (btn) => btn.className.includes('iconBtn') && !btn.className.includes('Danger')
            );
        moveButtons.forEach((btn) => expect(btn).toBeDisabled());
    });

    it('renders video thumbnails for each item', () => {
        render(<VideoPlaylistDetailView {...defaultProps()} />);

        // Hero image + 3 item thumbnails = 4 images total
        const images = screen.getAllByRole('img');
        expect(images).toHaveLength(4);
        expect(images[1]).toHaveAttribute(
            'src',
            'http://localhost:8000/api/v1/files/video-thumbnail/10?width=320&height=180'
        );
        expect(images[2]).toHaveAttribute(
            'src',
            'http://localhost:8000/api/v1/files/video-thumbnail/20?width=320&height=180'
        );
    });

    it('shows source kind text for each video item', () => {
        const playlist = createPlaylist({
            items: [
                createPlaylistItem(1, 10, 0, { source_kind: 'manual' }),
                createPlaylistItem(2, 20, 1, { source_kind: 'auto' }),
            ],
        });
        render(<VideoPlaylistDetailView {...defaultProps()} playlist={playlist} />);

        expect(screen.getByText(/Manual/)).toBeInTheDocument();
        expect(screen.getByText(/Auto/)).toBeInTheDocument();
    });

    it('renders empty list when playlist has no items', () => {
        const playlist = createPlaylist({ items: [], item_count: 0 });
        render(<VideoPlaylistDetailView {...defaultProps()} playlist={playlist} />);

        expect(screen.getByText('0 items')).toBeInTheDocument();
        // No video item buttons
        const images = screen.getAllByRole('img');
        // Only the hero image if cover_video_id is set
        expect(images).toHaveLength(1);
    });

    it('does not render hero image when no cover_video_id and no items', () => {
        const playlist = createPlaylist({
            cover_video_id: null,
            items: [],
            item_count: 0,
        });
        render(<VideoPlaylistDetailView {...defaultProps()} playlist={playlist} />);

        expect(screen.queryByRole('img')).toBeNull();
    });

    it('uppercases the classification in the eyebrow', () => {
        const playlist = createPlaylist({ classification: 'movie' });
        render(<VideoPlaylistDetailView {...defaultProps()} playlist={playlist} />);

        expect(screen.getByText('MOVIE')).toBeInTheDocument();
    });

    it('defaults classification to "personal" when empty', () => {
        const playlist = createPlaylist({ classification: '' as any });
        render(<VideoPlaylistDetailView {...defaultProps()} playlist={playlist} />);

        expect(screen.getByText('PERSONAL')).toBeInTheDocument();
    });

    it('renders edit playlist section heading', () => {
        render(<VideoPlaylistDetailView {...defaultProps()} />);

        expect(screen.getByText('Edit Playlist')).toBeInTheDocument();
    });

    it('initializes the name draft from playlist name', () => {
        render(<VideoPlaylistDetailView {...defaultProps()} />);

        const input = screen.getByDisplayValue('Test Playlist');
        expect(input).toBeInTheDocument();
    });

    it('sorts items by id when order_index is the same', () => {
        const playlist = createPlaylist({
            items: [
                createPlaylistItem(3, 30, 0),
                createPlaylistItem(1, 10, 0),
                createPlaylistItem(2, 20, 0),
            ],
        });
        const { container } = render(
            <VideoPlaylistDetailView {...defaultProps()} playlist={playlist} />
        );

        const titles = container.querySelectorAll('[class*="detailTitle"]');
        expect(titles[0]).toHaveTextContent('Video 10');
        expect(titles[1]).toHaveTextContent('Video 20');
        expect(titles[2]).toHaveTextContent('Video 30');
    });
});
