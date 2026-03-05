import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import { useImage, ImageProvider } from './imageProvider';
import { useInfiniteQuery } from '@tanstack/react-query';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
	apiBase: {
		get: jest.fn(),
	},
}));

jest.mock('@tanstack/react-query', () => ({
	useInfiniteQuery: jest.fn(),
}));

const mockedUseInfiniteQuery = useInfiniteQuery as jest.Mock;
const mockedApiGet = apiBase.get as jest.Mock;

function Consumer() {
	const ctx = useImage();
	return (
		<div>
			<span data-testid="count">{ctx.images.length}</span>
			<span data-testid="status">{ctx.status}</span>
			<span data-testid="group-by">{ctx.imageGroupBy}</span>
			<button type='button' onClick={() => ctx.setImageGroupBy('type')}>
				change-group
			</button>
		</div>
	);
}

describe('providers/imageProvider', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedUseInfiniteQuery.mockReturnValue({
			status: 'success',
			data: { pages: [{ items: [{ id: 1 }, { id: 2 }] }] },
			fetchNextPage: jest.fn(),
			hasNextPage: true,
			isFetchingNextPage: false,
		});
		mockedApiGet.mockResolvedValue({ data: { items: [], pagination: { has_next: false, page: 1 } } });
	});

	it('provides aggregated images', () => {
		render(
			<ImageProvider>
				<Consumer />
			</ImageProvider>,
		);

		expect(screen.getByTestId('count')).toHaveTextContent('2');
		expect(screen.getByTestId('status')).toHaveTextContent('success');
		expect(screen.getByTestId('group-by')).toHaveTextContent('date');
	});

	it('executes query function and next page resolver', async () => {
		render(
			<ImageProvider>
				<Consumer />
			</ImageProvider>,
		);

		const options = mockedUseInfiniteQuery.mock.calls[0][0];
		await options.queryFn({ pageParam: 3 });
		expect(mockedApiGet).toHaveBeenCalledWith('/files/images', {
			params: { page: 3, page_size: 200, group_by: 'date' },
		});
		await options.queryFn({});
		expect(mockedApiGet).toHaveBeenCalledWith('/files/images', {
			params: { page: 1, page_size: 200, group_by: 'date' },
		});
		expect(options.getNextPageParam({ pagination: { has_next: true, page: 4 } })).toBe(5);
		expect(options.getNextPageParam({ pagination: { has_next: false, page: 4 } })).toBeUndefined();
	});

	it('updates query key when grouping changes', async () => {
		const user = userEvent.setup();
		render(
			<ImageProvider>
				<Consumer />
			</ImageProvider>,
		);

		expect(mockedUseInfiniteQuery.mock.calls[0][0].queryKey).toEqual(['images', 'date']);
		await user.click(screen.getByRole('button', { name: 'change-group' }));
		expect(mockedUseInfiniteQuery.mock.calls[1][0].queryKey).toEqual(['images', 'type']);
	});

	it('throws when useImage is outside provider', () => {
		expect(() => render(<Consumer />)).toThrow('useImage must be used within an ImageContextProvider');
	});
});
