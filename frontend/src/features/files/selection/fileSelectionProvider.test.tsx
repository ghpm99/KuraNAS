import { fireEvent, render, screen } from '@testing-library/react';
import FileSelectionProvider from './fileSelectionProvider';
import { useFileSelectionContext } from './fileSelectionContext';
import { createTestFile } from './testFileFactory';

const SelectionProbe = () => {
    const { selectedCount, toggle } = useFileSelectionContext();
    return (
        <div>
            <span data-testid="count">{selectedCount}</span>
            <button onClick={() => toggle(createTestFile(1))}>select</button>
        </div>
    );
};

describe('FileSelectionProvider', () => {
    it('renders children with an empty selection', () => {
        render(
            <FileSelectionProvider scopeKey="root">
                <SelectionProbe />
            </FileSelectionProvider>
        );

        expect(screen.getByTestId('count')).toHaveTextContent('0');
    });

    it('exposes an inert selection outside the provider', () => {
        render(<SelectionProbe />);

        fireEvent.click(screen.getByText('select'));
        expect(screen.getByTestId('count')).toHaveTextContent('0');
    });

    it('clears the selection on Escape and ignores other keys', () => {
        render(
            <FileSelectionProvider scopeKey="root">
                <SelectionProbe />
            </FileSelectionProvider>
        );

        fireEvent.click(screen.getByText('select'));
        expect(screen.getByTestId('count')).toHaveTextContent('1');

        fireEvent.keyDown(document, { key: 'Enter' });
        expect(screen.getByTestId('count')).toHaveTextContent('1');

        fireEvent.keyDown(document, { key: 'Escape' });
        expect(screen.getByTestId('count')).toHaveTextContent('0');
    });
});
