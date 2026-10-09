import { fireEvent, render, screen } from '@testing-library/react';
import GlobalSearchDialog from './GlobalSearchDialog';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string>) =>
            params ? `${key}:${params.query}` : key,
    }),
}));

const baseProps = {
    open: true,
    query: '',
    sections: [],
    isFetching: false,
    activeItemId: '',
    shortcut: 'Ctrl+K',
    showEmptyState: false,
    onClose: jest.fn(),
    onQueryChange: jest.fn(),
    onInputKeyDown: jest.fn(),
    onItemHover: jest.fn(),
    onItemSelect: jest.fn(),
};

describe('GlobalSearchDialog recent searches and fuzzy notice', () => {
    it('renders without the optional props', () => {
        render(<GlobalSearchDialog {...baseProps} />);

        expect(screen.queryByText('GLOBAL_SEARCH_SECTION_RECENT')).not.toBeInTheDocument();
        expect(screen.queryByText('GLOBAL_SEARCH_FUZZY_NOTICE')).not.toBeInTheDocument();
    });

    it('lists recent searches and wires select, remove and clear', () => {
        const onSelect = jest.fn();
        const onRemove = jest.fn();
        const onClear = jest.fn();
        render(
            <GlobalSearchDialog
                {...baseProps}
                recentSearches={['ferias', 'contrato']}
                onRecentSearchSelect={onSelect}
                onRecentSearchRemove={onRemove}
                onRecentSearchesClear={onClear}
            />
        );

        fireEvent.click(screen.getByText('contrato'));
        fireEvent.click(screen.getByLabelText('GLOBAL_SEARCH_RECENT_REMOVE:ferias'));
        fireEvent.click(screen.getByText('GLOBAL_SEARCH_RECENT_CLEAR'));

        expect(onSelect).toHaveBeenCalledWith('contrato');
        expect(onRemove).toHaveBeenCalledWith('ferias');
        expect(onClear).toHaveBeenCalled();
    });

    it('hides recent searches while a query is typed', () => {
        render(<GlobalSearchDialog {...baseProps} query="fe" recentSearches={['ferias']} />);

        expect(screen.queryByText('GLOBAL_SEARCH_SECTION_RECENT')).not.toBeInTheDocument();
    });

    it('shows the fuzzy notice only for fuzzy results', () => {
        const { rerender } = render(<GlobalSearchDialog {...baseProps} isFuzzyResult />);
        expect(screen.getByText('GLOBAL_SEARCH_FUZZY_NOTICE')).toBeInTheDocument();

        rerender(<GlobalSearchDialog {...baseProps} isFuzzyResult={false} />);
        expect(screen.queryByText('GLOBAL_SEARCH_FUZZY_NOTICE')).not.toBeInTheDocument();
    });

    it('highlights the query in result titles', () => {
        const { container } = render(
            <GlobalSearchDialog
                {...baseProps}
                query="ferias"
                sections={[
                    {
                        id: 'files',
                        title: 'Files',
                        items: [
                            {
                                id: 'file-1',
                                kind: 'file',
                                label: 'Férias 2024.pdf',
                                description: '/docs',
                                onSelect: jest.fn(),
                            },
                        ],
                    },
                ]}
            />
        );

        expect(container.ownerDocument.querySelector('mark')?.textContent).toBe('Férias');
    });
});
