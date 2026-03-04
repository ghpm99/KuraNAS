import { render, screen } from '@testing-library/react';
import React from 'react';
import ImageContent from './imageContent';

const mockUseImage = jest.fn();
const mockRef = jest.fn();
const mockUseIntersectionObserver = jest.fn();

jest.mock('../hooks/imageProvider/imageProvider', () => ({ useImage: () => mockUseImage() }));
jest.mock('../hooks/IntersectionObserver/useIntersectionObserver', () => ({
	useIntersectionObserver: (...args: any[]) => mockUseIntersectionObserver(...args),
}));

describe('imageContent', () => {
	beforeEach(() => {
		mockUseIntersectionObserver.mockImplementation(() => ({ ref: mockRef }));
	});

	it('renders images and loading/end states', () => {
		mockUseImage.mockReturnValue({
			images: [{ id: 1, name: 'img1', format: '.jpg', size: 1024 }],
			fetchNextPage: jest.fn(),
			hasNextPage: false,
			isFetchingNextPage: true,
		});
		render(<ImageContent />);
		expect(screen.getByAltText('img1')).toBeInTheDocument();
		expect(screen.getByRole('progressbar')).toBeInTheDocument();
		expect(screen.getByText('Todas as imagens carregadas')).toBeInTheDocument();
	});

	it('triggers infinite load on intersect when enabled', () => {
		const fetchNextPage = jest.fn();
		let optionsRef: any;
		mockUseIntersectionObserver.mockImplementation((options: any) => {
			optionsRef = options;
			return { ref: mockRef };
		});
		mockUseImage.mockReturnValue({
			images: [{ id: 1, name: 'img1', format: '.jpg', size: 1024 }],
			fetchNextPage,
			hasNextPage: true,
			isFetchingNextPage: false,
		});
		render(<ImageContent />);
		optionsRef.onIntersect();
		expect(fetchNextPage).toHaveBeenCalled();
		expect(screen.queryByText('Todas as imagens carregadas')).not.toBeInTheDocument();
	});
});
