import { render, screen } from '@testing-library/react';
import HighlightedText from './HighlightedText';

describe('HighlightedText', () => {
    it('renders without props', () => {
        const { container } = render(<HighlightedText />);

        expect(container).toBeEmptyDOMElement();
    });

    it('renders plain text when the query is empty', () => {
        const { container } = render(<HighlightedText text="Relatorio" />);

        expect(container.querySelector('mark')).toBeNull();
        expect(screen.getByText('Relatorio')).toBeInTheDocument();
    });

    it('highlights ignoring case and accents while keeping original characters', () => {
        const { container } = render(
            <HighlightedText text="Férias de São João" query="ferias sao" />
        );

        const marks = Array.from(container.querySelectorAll('mark')).map(
            (mark) => mark.textContent
        );
        expect(marks).toEqual(['Férias', 'São']);
        expect(container.textContent).toBe('Férias de São João');
    });

    it('highlights accented characters when the query carries the accent', () => {
        const { container } = render(<HighlightedText text="Cafe da manha" query="café" />);

        expect(container.querySelector('mark')?.textContent).toBe('Cafe');
    });

    it('highlights every occurrence of a term', () => {
        const { container } = render(<HighlightedText text="aba aba" query="ab" />);

        expect(container.querySelectorAll('mark')).toHaveLength(2);
    });
});
