import { fireEvent, render, screen } from '@testing-library/react';
import FileCard from './fileCard';

describe('components/fileCard', () => {
	it('renders card with thumbnail, metadata and handlers', () => {
		const onClick = jest.fn();
		const onClickStar = jest.fn();
		render(
			<FileCard
				title='Photo'
				metadata='jpg - 1 MB'
				thumbnail='/photo.jpg'
				onClick={onClick}
				starred
				onClickStar={onClickStar}
			/>,
		);

		expect(screen.getByText('Photo')).toBeInTheDocument();
		expect(screen.getByText('jpg - 1 MB')).toBeInTheDocument();
		fireEvent.click(screen.getByText('Photo'));
		expect(onClick).toHaveBeenCalled();
		fireEvent.click(screen.getAllByRole('button')[1]!);
		expect(onClickStar).toHaveBeenCalled();
	});

	it('uses placeholder thumbnail fallback when image is empty', () => {
		render(
			<FileCard title='No Image' metadata='meta' thumbnail='' onClick={jest.fn()} />,
		);
		expect(screen.getByAltText('No Image')).toHaveAttribute('src', '/placeholder.svg');
	});
});
