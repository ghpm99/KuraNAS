import { fireEvent, render, screen } from '@testing-library/react';
import VideoContextDetailView from './VideoContextDetailView';
import VideoSeriesDetailView from './VideoSeriesDetailView';

jest.mock('@/service/apiUrl', () => ({
	getApiV1BaseUrl: () => 'http://localhost:8000/v1',
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string, params?: Record<string, string | number>) => {
			if (key === 'VIDEO_DETAIL_COLLECTION_META') return `${params?.count ?? 0} itens`;
			if (key === 'VIDEO_DETAIL_SERIES_PROGRESS') return `${params?.completed ?? 0} de ${params?.count ?? 0} concluidos`;
			if (key === 'VIDEO_DETAIL_SEASON_LABEL') return `Temporada ${params?.season ?? ''}`.trim();
			if (key === 'VIDEO_DETAIL_SEASON_DESCRIPTION') return `${params?.count ?? 0} episodios`;
			const map: Record<string, string> = {
				VIDEO_BACK_TO_VIDEOS: 'Voltar para videos',
				VIDEO_DETAIL_RESUME_ACTION: 'Retomar agora',
				VIDEO_DETAIL_COLLECTION_TOTAL: 'Total',
				VIDEO_DETAIL_COLLECTION_COMPLETED: 'Concluidos',
				VIDEO_DETAIL_COLLECTION_PENDING: 'Pendentes',
				VIDEO_DETAIL_COLLECTION_ITEMS: 'Itens da colecao',
				VIDEO_DETAIL_COLLECTION_ITEMS_DESCRIPTION: 'Descricao',
				VIDEO_DETAIL_SERIES_EYEBROW: 'Serie',
				VIDEO_PLAY: 'Reproduzir',
				VIDEO_SOURCE_AUTO: 'auto',
				VIDEO_SOURCE_MANUAL: 'manual',
				VIDEO_STATUS_NOT_STARTED: 'Nao iniciado',
				VIDEO_STATUS_IN_PROGRESS: 'Em andamento',
				VIDEO_STATUS_COMPLETED: 'Concluido',
			};

			return map[key] ?? key;
		},
	}),
}));

describe('video detail views', () => {
	it('renders context detail fallback without resume item', () => {
		const onBack = jest.fn();
		const onOpenVideo = jest.fn();

		render(
			<VideoContextDetailView
				playlist={{
					id: 9,
					name: 'Loose Clips',
					type: 'custom',
					source_path: '/clips',
					is_hidden: false,
					is_auto: true,
					group_mode: 'single',
					classification: '' as unknown as 'personal',
					item_count: 0,
					cover_video_id: null,
					created_at: '2026-03-14T00:00:00Z',
					updated_at: '2026-03-14T00:00:00Z',
					last_played_at: null,
					items: [],
				}}
				onBack={onBack}
				onOpenVideo={onOpenVideo}
			/>,
		);

		expect(screen.getByText('PERSONAL')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: /Reproduzir/i })).toBeDisabled();

		fireEvent.click(screen.getByRole('button', { name: /Voltar para videos/i }));
		expect(onBack).toHaveBeenCalled();
	});

	it('renders context detail with resume item and opens the focused video', () => {
		const onOpenVideo = jest.fn();

		render(
			<VideoContextDetailView
				playlist={{
					id: 10,
					name: 'Movie Night',
					type: 'movie',
					source_path: '/movies',
					is_hidden: false,
					is_auto: true,
					group_mode: 'single',
					classification: 'movie',
					item_count: 1,
					cover_video_id: 91,
					created_at: '2026-03-14T00:00:00Z',
					updated_at: '2026-03-14T00:00:00Z',
					last_played_at: null,
					items: [
						{
							id: 1,
							order_index: 0,
							source_kind: 'auto',
							status: 'in_progress',
							progress_pct: 40,
							video: {
								id: 91,
								name: 'Movie Night.mkv',
								path: '/movies/movie-night.mkv',
								parent_path: '/movies',
								format: 'mkv',
								size: 120,
							},
						},
					],
				}}
				onBack={jest.fn()}
				onOpenVideo={onOpenVideo}
			/>,
		);

		fireEvent.click(screen.getByRole('button', { name: /Retomar agora/i }));
		expect(onOpenVideo).toHaveBeenCalledWith(91);
	});

	it('renders series detail with multiple seasons and toggles the selected season', () => {
		render(
			<VideoSeriesDetailView
				playlist={{
					id: 11,
					name: 'Showcase',
					type: 'series',
					source_path: '/series/showcase',
					is_hidden: false,
					is_auto: true,
					group_mode: 'prefix',
					classification: 'series',
					item_count: 2,
					cover_video_id: 101,
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
								id: 101,
								name: 'Showcase S01E01.mkv',
								path: '/series/showcase/s01e01.mkv',
								parent_path: '/series/showcase',
								format: 'mkv',
								size: 80,
							},
						},
						{
							id: 2,
							order_index: 1,
							source_kind: 'auto',
							status: 'not_started',
							progress_pct: 0,
							video: {
								id: 102,
								name: 'Showcase S02E01.mkv',
								path: '/series/showcase/s02e01.mkv',
								parent_path: '/series/showcase',
								format: 'mkv',
								size: 82,
							},
						},
					],
				}}
				onBack={jest.fn()}
				onOpenVideo={jest.fn()}
			/>,
		);

		expect(screen.getByRole('button', { name: 'Temporada 2' })).toBeInTheDocument();
		fireEvent.click(screen.getByRole('button', { name: 'Temporada 2' }));
		expect(screen.getByText('1 de 2 concluidos')).toBeInTheDocument();
	});

	it('falls back to a single synthetic season when episode numbering is missing', () => {
		render(
			<VideoSeriesDetailView
				playlist={{
					id: 12,
					name: 'Unsorted Series',
					type: 'series',
					source_path: '/series/unsorted',
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
							id: 3,
							order_index: 0,
							source_kind: 'manual',
							status: 'not_started',
							progress_pct: 0,
							video: {
								id: 103,
								name: 'Pilot.mkv',
								path: '/series/unsorted/pilot.mkv',
								parent_path: '/series/unsorted',
								format: 'mkv',
								size: 70,
							},
						},
					],
				}}
				onBack={jest.fn()}
				onOpenVideo={jest.fn()}
			/>,
		);

		expect(screen.getByText('Temporada 1')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: /Reproduzir/i })).toBeInTheDocument();
	});
});
