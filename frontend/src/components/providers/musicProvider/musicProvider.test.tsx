import { act, render, screen } from '@testing-library/react';
import React from 'react';
import { MusicProvider, useMusic } from './musicProvider';
import { useInfiniteQuery } from '@tanstack/react-query';
import { apiBase } from '@/service';
import { useIntersectionObserver } from '../IntersectionObserver/useIntersectionObserver';

jest.mock('@/service', () => ({
	apiBase: {
		get: jest.fn(),
	},
}));

jest.mock('../IntersectionObserver/useIntersectionObserver', () => ({
	useIntersectionObserver: jest.fn(() => ({ ref: jest.fn() })),
}));

jest.mock('@tanstack/react-query', () => ({
	useInfiniteQuery: jest.fn(),
}));

const mockedUseInfiniteQuery = useInfiniteQuery as jest.Mock;
const mockedApiGet = apiBase.get as jest.Mock;
const mockedUseIntersectionObserver = useIntersectionObserver as jest.Mock;
const mockFetchNextPage = jest.fn();

function Consumer() {
	const ctx = useMusic();
	return (
		<div>
			<span data-testid="count">{ctx.music.length}</span>
			<span data-testid="status">{ctx.status}</span>
			<button onClick={() => ctx.setCurrentView('artists')}>set-view</button>
			<span data-testid="view">{ctx.currentView}</span>
		</div>
	);
}

describe('providers/musicProvider', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedUseInfiniteQuery.mockReturnValue({
			status: 'success',
			data: { pages: [{ items: [{ id: 1 }] }] },
			fetchNextPage: mockFetchNextPage,
			hasNextPage: true,
			isFetchingNextPage: false,
		});
		mockedApiGet.mockResolvedValue({ data: { items: [], pagination: { has_next: false, page: 1 } } });
		mockedUseIntersectionObserver.mockImplementation(() => ({ ref: jest.fn() }));
	});

	it('provides music context and updates current view', () => {
		render(
			<MusicProvider>
				<Consumer />
			</MusicProvider>,
		);

		expect(screen.getByTestId('count')).toHaveTextContent('1');
		expect(screen.getByTestId('status')).toHaveTextContent('success');
		expect(screen.getByTestId('view')).toHaveTextContent('all');
		act(() => {
			screen.getByText('set-view').click();
		});
		expect(screen.getByTestId('view')).toHaveTextContent('artists');
	});

	it('executes query function and next page resolver', async () => {
		render(
			<MusicProvider>
				<Consumer />
			</MusicProvider>,
		);

		const options = mockedUseInfiniteQuery.mock.calls[0][0];
		await options.queryFn({ pageParam: 2 });
		expect(mockedApiGet).toHaveBeenCalledWith('/files/music', {
			params: { page: 2, page_size: 200 },
		});
		expect(options.getNextPageParam({ pagination: { has_next: true, page: 9 } })).toBe(10);
		expect(options.getNextPageParam({ pagination: { has_next: false, page: 9 } })).toBeUndefined();
	});

	it('throws when useMusic is outside provider', () => {
		expect(() => render(<Consumer />)).toThrow('useMusic must be used within a MusicContextProvider');
	});

	it('triggers intersection callback only when pagination allows', () => {
		let observerArgs: any;
		mockedUseIntersectionObserver.mockImplementation((args: any) => {
			observerArgs = args;
			return { ref: jest.fn() };
		});
		mockedUseInfiniteQuery.mockReturnValue({
			status: 'success',
			data: { pages: [{ items: [{ id: 1 }] }] },
			fetchNextPage: mockFetchNextPage,
			hasNextPage: true,
			isFetchingNextPage: false,
		});
		const { rerender } = render(
			<MusicProvider>
				<Consumer />
			</MusicProvider>,
		);

		act(() => {
			observerArgs.onIntersect();
		});
		expect(mockFetchNextPage).toHaveBeenCalled();

		mockedUseInfiniteQuery.mockReturnValue({
			status: 'success',
			data: { pages: [{ items: [{ id: 1 }] }] },
			fetchNextPage: mockFetchNextPage,
			hasNextPage: false,
			isFetchingNextPage: false,
		});
		rerender(
			<MusicProvider>
				<Consumer />
			</MusicProvider>,
		);
		act(() => {
			observerArgs.onIntersect();
		});
		expect(mockFetchNextPage).toHaveBeenCalledTimes(1);
	});
});
