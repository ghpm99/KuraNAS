import { render, screen, fireEvent } from '@testing-library/react';
import ImageCategoryTabs from './ImageCategoryTabs';

jest.mock('../../i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string) => key,
	}),
}));

type TestCategory = 'photos' | 'screenshots' | 'wallpapers';

const labels: Record<TestCategory, string> = {
	photos: 'Photos',
	screenshots: 'Screenshots',
	wallpapers: 'Wallpapers',
};

const counts: Record<TestCategory, number> = {
	photos: 42,
	screenshots: 7,
	wallpapers: 13,
};

describe('ImageCategoryTabs', () => {
	it('renders a button for each category with label and count', () => {
		render(
			<ImageCategoryTabs<TestCategory>
				activeCategory='photos'
				labels={labels}
				counts={counts}
				onSelect={jest.fn()}
			/>,
		);

		expect(screen.getByText('Photos')).toBeInTheDocument();
		expect(screen.getByText('42')).toBeInTheDocument();
		expect(screen.getByText('Screenshots')).toBeInTheDocument();
		expect(screen.getByText('7')).toBeInTheDocument();
		expect(screen.getByText('Wallpapers')).toBeInTheDocument();
		expect(screen.getByText('13')).toBeInTheDocument();
	});

	it('renders a tablist with the correct aria label', () => {
		render(
			<ImageCategoryTabs<TestCategory>
				activeCategory='photos'
				labels={labels}
				counts={counts}
				onSelect={jest.fn()}
			/>,
		);

		expect(screen.getByRole('tablist', { name: 'IMAGES_CATEGORIES_ARIA' })).toBeInTheDocument();
	});

	it('applies active class to the active category button', () => {
		const { container } = render(
			<ImageCategoryTabs<TestCategory>
				activeCategory='screenshots'
				labels={labels}
				counts={counts}
				onSelect={jest.fn()}
			/>,
		);

		const buttons = container.querySelectorAll('button');
		// The active button (screenshots, index 1) should have the active class
		expect(buttons[1]?.className).toContain('categoryPillActive');
		// Other buttons should not have the active class
		expect(buttons[0]?.className).not.toContain('categoryPillActive');
		expect(buttons[2]?.className).not.toContain('categoryPillActive');
	});

	it('calls onSelect with the correct category when a button is clicked', () => {
		const onSelect = jest.fn();
		render(
			<ImageCategoryTabs<TestCategory>
				activeCategory='photos'
				labels={labels}
				counts={counts}
				onSelect={onSelect}
			/>,
		);

		fireEvent.click(screen.getByText('Screenshots'));
		expect(onSelect).toHaveBeenCalledWith('screenshots');

		fireEvent.click(screen.getByText('Wallpapers'));
		expect(onSelect).toHaveBeenCalledWith('wallpapers');
	});

	it('calls onSelect when clicking the already active category', () => {
		const onSelect = jest.fn();
		render(
			<ImageCategoryTabs<TestCategory>
				activeCategory='photos'
				labels={labels}
				counts={counts}
				onSelect={onSelect}
			/>,
		);

		fireEvent.click(screen.getByText('Photos'));
		expect(onSelect).toHaveBeenCalledWith('photos');
	});
});
