import { fireEvent, render, screen } from '@testing-library/react';
import GlobalSearchDialog from './GlobalSearchDialog';
import type { SearchDialogItem, SearchDialogSection } from './useGlobalSearchProvider';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

const createItem = (overrides: Partial<SearchDialogItem> = {}): SearchDialogItem => ({
	id: 'item-1',
	kind: 'file',
	label: 'Test Item',
	description: 'Item description',
	onSelect: jest.fn(),
	...overrides,
});

const createSection = (overrides: Partial<SearchDialogSection> = {}): SearchDialogSection => ({
	id: 'section-1',
	title: 'Section',
	items: [createItem()],
	...overrides,
});

const defaultProps = () => ({
	open: true,
	query: '',
	sections: [] as SearchDialogSection[],
	isFetching: false,
	activeItemId: '',
	shortcut: 'Ctrl+K',
	showEmptyState: false,
	onClose: jest.fn(),
	onQueryChange: jest.fn(),
	onInputKeyDown: jest.fn(),
	onItemHover: jest.fn(),
	onItemSelect: jest.fn(),
});

describe('GlobalSearchDialog', () => {
	describe('getItemIcon', () => {
		const kinds = ['folder', 'artist', 'album', 'playlist', 'video', 'image', 'action', 'file'] as const;

		it.each(kinds)('renders an icon for kind "%s"', (kind) => {
			const item = createItem({ id: `icon-${kind}`, kind, label: kind });
			const section = createSection({ items: [item] });
			const props = defaultProps();
			render(<GlobalSearchDialog {...props} sections={[section]} />);
			expect(screen.getByText(kind)).toBeInTheDocument();
		});
	});

	it('renders sections with items', () => {
		const items = [
			createItem({ id: 'a', label: 'Alpha', description: 'desc-a' }),
			createItem({ id: 'b', label: 'Beta', description: 'desc-b', meta: 'mp4' }),
		];
		const section = createSection({ title: 'Results', items });
		const props = defaultProps();
		render(<GlobalSearchDialog {...props} sections={[section]} />);

		expect(screen.getByText('Results')).toBeInTheDocument();
		expect(screen.getByText('Alpha')).toBeInTheDocument();
		expect(screen.getByText('Beta')).toBeInTheDocument();
		expect(screen.getByText('mp4')).toBeInTheDocument();
	});

	it('renders empty state when showEmptyState is true', () => {
		const props = defaultProps();
		render(<GlobalSearchDialog {...props} showEmptyState />);

		expect(screen.getByText('GLOBAL_SEARCH_EMPTY_TITLE')).toBeInTheDocument();
		expect(screen.getByText('GLOBAL_SEARCH_EMPTY_DESCRIPTION')).toBeInTheDocument();
	});

	it('does not render empty state when showEmptyState is false', () => {
		const props = defaultProps();
		render(<GlobalSearchDialog {...props} showEmptyState={false} />);

		expect(screen.queryByText('GLOBAL_SEARCH_EMPTY_TITLE')).not.toBeInTheDocument();
	});

	it('shows loading spinner when isFetching is true', () => {
		const props = defaultProps();
		render(<GlobalSearchDialog {...props} isFetching />);

		expect(screen.getByRole('progressbar')).toBeInTheDocument();
		expect(screen.queryByText('Ctrl+K')).not.toBeInTheDocument();
	});

	it('shows shortcut when not fetching', () => {
		const props = defaultProps();
		render(<GlobalSearchDialog {...props} shortcut='Cmd+K' />);

		expect(screen.getByText('Cmd+K')).toBeInTheDocument();
	});

	it('calls onQueryChange when input value changes', () => {
		const props = defaultProps();
		render(<GlobalSearchDialog {...props} />);

		const input = screen.getByRole('textbox');
		fireEvent.change(input, { target: { value: 'hello' } });

		expect(props.onQueryChange).toHaveBeenCalledWith('hello');
	});

	it('calls onInputKeyDown on key press', () => {
		const props = defaultProps();
		render(<GlobalSearchDialog {...props} />);

		const input = screen.getByRole('textbox');
		fireEvent.keyDown(input, { key: 'ArrowDown' });

		expect(props.onInputKeyDown).toHaveBeenCalled();
	});

	it('calls onItemHover on mouse enter', () => {
		const item = createItem({ id: 'hover-item', label: 'Hover' });
		const section = createSection({ items: [item] });
		const props = defaultProps();
		render(<GlobalSearchDialog {...props} sections={[section]} />);

		fireEvent.mouseEnter(screen.getByRole('button', { name: /Hover/i }));

		expect(props.onItemHover).toHaveBeenCalledWith('hover-item');
	});

	it('calls onItemSelect on click', () => {
		const item = createItem({ id: 'click-item', label: 'Clickable' });
		const section = createSection({ items: [item] });
		const props = defaultProps();
		render(<GlobalSearchDialog {...props} sections={[section]} />);

		fireEvent.click(screen.getByRole('button', { name: /Clickable/i }));

		expect(props.onItemSelect).toHaveBeenCalledWith(item);
	});

	it('applies active class to the active item', () => {
		const item = createItem({ id: 'active-1', label: 'Active Item' });
		const section = createSection({ items: [item] });
		const props = defaultProps();
		render(<GlobalSearchDialog {...props} sections={[section]} activeItemId='active-1' />);

		const button = screen.getByRole('button', { name: /Active Item/i });
		expect(button.className).toContain('itemActive');
	});

	it('does not apply active class to non-active items', () => {
		const item = createItem({ id: 'other-1', label: 'Other Item' });
		const section = createSection({ items: [item] });
		const props = defaultProps();
		render(<GlobalSearchDialog {...props} sections={[section]} activeItemId='different-id' />);

		const button = screen.getByRole('button', { name: /Other Item/i });
		expect(button.className).not.toContain('itemActive');
	});

	it('does not render meta span when item has no meta', () => {
		const item = createItem({ id: 'no-meta', label: 'No Meta', meta: undefined });
		const section = createSection({ items: [item] });
		const props = defaultProps();
		const { container } = render(<GlobalSearchDialog {...props} sections={[section]} />);

		const metaSpans = container.querySelectorAll('[class*="itemMeta"]');
		expect(metaSpans).toHaveLength(0);
	});
});
