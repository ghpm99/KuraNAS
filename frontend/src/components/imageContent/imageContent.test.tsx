import { fireEvent, render, screen } from '@testing-library/react';
import ImageContent from './imageContent';

const mockUseImage = jest.fn();
const mockRef = jest.fn();
const mockUseIntersectionObserver = jest.fn();

jest.mock('../hooks/imageProvider/imageProvider', () => ({ useImage: () => mockUseImage() }));
jest.mock('../hooks/IntersectionObserver/useIntersectionObserver', () => ({
	useIntersectionObserver: (...args: any[]) => mockUseIntersectionObserver(...args),
}));
jest.mock('@/service/apiUrl', () => ({
	getApiV1BaseUrl: () => '/api/v1',
}));

describe('imageContent', () => {
	beforeEach(() => {
		mockUseIntersectionObserver.mockImplementation(() => ({ ref: mockRef }));
	});

	it('renders grouped gallery and opens viewer with details', () => {
		mockUseImage.mockReturnValue({
			images: [
				{
					id: 1,
					name: 'img1',
					path: '/photos',
					format: '.jpg',
					size: 1024,
					created_at: '2025-01-10T10:00:00Z',
					updated_at: '2025-01-10T10:00:00Z',
					metadata: { width: 1600, height: 900, make: 'Sony', model: 'A7' },
				},
			],
			imageGroupBy: 'date',
			setImageGroupBy: jest.fn(),
			fetchNextPage: jest.fn(),
			hasNextPage: false,
			isFetchingNextPage: true,
		});

		render(<ImageContent />);

		expect(screen.getByText('Galeria de fotos')).toBeInTheDocument();
		expect(screen.getByText('Todas as imagens carregadas')).toBeInTheDocument();
		expect(screen.getByRole('progressbar')).toBeInTheDocument();

		fireEvent.click(screen.getByRole('button', { name: /abrir img1/i }));
		expect(screen.getByRole('dialog')).toBeInTheDocument();
		expect(screen.getByText('Detalhes')).toBeInTheDocument();

		fireEvent.click(screen.getByRole('button', { name: 'Fechar visualizador' }));
		expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
	});

	it('triggers infinite load on intersect when enabled', () => {
		const fetchNextPage = jest.fn();
		let optionsRef: any;
		mockUseIntersectionObserver.mockImplementation((options: any) => {
			optionsRef = options;
			return { ref: mockRef };
		});

		mockUseImage.mockReturnValue({
			images: [
				{
					id: 1,
					name: 'img1',
					path: '/photos',
					format: '.jpg',
					size: 1024,
					created_at: '2025-01-10T10:00:00Z',
					updated_at: '2025-01-10T10:00:00Z',
				},
			],
			imageGroupBy: 'date',
			setImageGroupBy: jest.fn(),
			fetchNextPage,
			hasNextPage: true,
			isFetchingNextPage: false,
		});

		render(<ImageContent />);
		optionsRef.onIntersect();

		expect(fetchNextPage).toHaveBeenCalled();
		expect(screen.queryByText('Todas as imagens carregadas')).not.toBeInTheDocument();
	});

	it('changes grouping through selector', () => {
		const setImageGroupBy = jest.fn();
		mockUseImage.mockReturnValue({
			images: [
				{
					id: 1,
					name: 'alpha.jpg',
					path: '/photos',
					format: '.jpg',
					size: 1024,
					created_at: '2025-01-10T10:00:00Z',
					updated_at: '2025-01-10T10:00:00Z',
				},
			],
			imageGroupBy: 'date',
			setImageGroupBy,
			fetchNextPage: jest.fn(),
			hasNextPage: false,
			isFetchingNextPage: false,
		});

		render(<ImageContent />);
		fireEvent.change(screen.getByLabelText('Agrupar imagens por'), { target: { value: 'type' } });

		expect(setImageGroupBy).toHaveBeenCalledWith('type');
	});
});
