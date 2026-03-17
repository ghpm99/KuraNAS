import { renderHook } from '@testing-library/react';
import { useVideoPlaylistDetail } from './useVideoPlaylistDetail';

describe('useVideoPlaylistDetail', () => {
    it('parses episode-only naming conventions into season 1 sequence labels', () => {
        const { result } = renderHook(() =>
            useVideoPlaylistDetail({
                id: 1,
                name: 'Episodes',
                type: 'series',
                source_path: '/series/episodes',
                is_hidden: false,
                is_auto: true,
                group_mode: 'prefix',
                classification: 'series',
                item_count: 1,
                cover_video_id: null,
                created_at: '2026-03-14T00:00:00Z',
                updated_at: '2026-03-14T00:00:00Z',
                last_played_at: null,
                items: [
                    {
                        id: 1,
                        order_index: 0,
                        source_kind: 'auto',
                        status: 'not_started',
                        progress_pct: 0,
                        video: {
                            id: 9,
                            name: 'Pilot Episode 7.mkv',
                            path: '/series/episodes/pilot-episode-7.mkv',
                            parent_path: '/series/episodes',
                            format: 'mkv',
                            size: 55,
                        },
                    },
                ],
            })
        );

        expect(result.current.orderedItems[0]?.sequenceLabel).toBe('S01E07');
        expect(result.current.groupedSeasons[0]?.label).toBe('1');
    });

    it('keeps non-episodic files in a flat collection and picks the first pending item to resume', () => {
        const { result } = renderHook(() =>
            useVideoPlaylistDetail({
                id: 2,
                name: 'Flat Clips',
                type: 'custom',
                source_path: '/clips',
                is_hidden: false,
                is_auto: true,
                group_mode: 'single',
                classification: 'clip',
                item_count: 2,
                cover_video_id: null,
                created_at: '2026-03-14T00:00:00Z',
                updated_at: '2026-03-14T00:00:00Z',
                last_played_at: null,
                items: [
                    {
                        id: 1,
                        order_index: 0,
                        source_kind: 'auto',
                        status: 'completed',
                        progress_pct: 100,
                        video: {
                            id: 10,
                            name: 'clip-one.mp4',
                            path: '/clips/clip-one.mp4',
                            parent_path: '/clips',
                            format: 'mp4',
                            size: 10,
                        },
                    },
                    {
                        id: 2,
                        order_index: 1,
                        source_kind: 'auto',
                        status: 'not_started',
                        progress_pct: 0,
                        video: {
                            id: 11,
                            name: 'clip-two.mp4',
                            path: '/clips/clip-two.mp4',
                            parent_path: '/clips',
                            format: 'mp4',
                            size: 10,
                        },
                    },
                ],
            })
        );

        expect(result.current.groupedSeasons).toHaveLength(0);
        expect(result.current.resumeItem?.video.id).toBe(11);
        expect(result.current.orderedItems[0]?.displayTitle).toBe('clip one');
    });
});
