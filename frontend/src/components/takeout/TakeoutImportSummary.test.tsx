import { apiBase } from '@/service';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import type { TakeoutImport } from '@/types/takeout';
import { screen, waitFor } from '@testing-library/react';
import TakeoutImportSummary from './TakeoutImportSummary';

jest.mock('@/service', () => {
	const actual = jest.requireActual('@/service');
	return { ...actual, apiBase: { ...actual.apiBase, get: jest.fn() } };
});

const mockedApiGet = apiBase.get as unknown as jest.Mock;

const completedImport: TakeoutImport = {
	id: 7,
	job_id: 3,
	file_name: 'takeout.zip',
	size: 100,
	state: 'completed',
	total_entries: 10,
	processed_entries: 10,
	imported: 6,
	already_existing: 2,
	ignored: 1,
	failed_entries: 1,
	error: '',
	created_at: '2026-11-25T10:00:00Z',
	ended_at: null,
};

describe('TakeoutImportSummary', () => {
	beforeEach(() => {
		mockedApiGet.mockReset();
	});

	it('renders without backend with the counts and links to the unfiltered folder views', async () => {
		mockedApiGet.mockRejectedValue(new Error('offline'));

		renderWithoutBackend(<TakeoutImportSummary takeoutImport={completedImport} />);

		await waitFor(() => expect(mockedApiGet).toHaveBeenCalled());
		const links = screen.getAllByRole('link');
		expect(links.map((link) => link.getAttribute('href'))).toEqual([
			'/images/folders',
			'/videos/folders',
		]);
	});

	it('links to the takeout folder of the image and video libraries', async () => {
		mockedApiGet.mockImplementation(
			async (url: string, config?: { params?: { path?: string } }) => {
				if (url === '/libraries') {
					return {
						data: [
							{ category: 'images', path: '/nas/Imagens' },
							{ category: 'videos', path: '/nas/Videos' },
						],
					};
				}
				const logicalPathByDiskPath: Record<string, string> = {
					'/nas/Imagens': '/Imagens',
					'/nas/Videos': '/Videos',
				};
				return { data: { path: logicalPathByDiskPath[config?.params?.path ?? ''] } };
			}
		);

		renderWithoutBackend(<TakeoutImportSummary takeoutImport={completedImport} />);

		const [photosLink, videosLink] = screen.getAllByRole('link');
		await waitFor(() =>
			expect(photosLink?.getAttribute('href')).toBe(
				'/images/folders?folder=%2FImagens%2Ftakeout'
			)
		);
		expect(videosLink?.getAttribute('href')).toBe('/videos/folders?folder=%2FVideos%2Ftakeout');
	});
});
