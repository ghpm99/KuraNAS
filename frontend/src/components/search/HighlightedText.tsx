import { useMemo } from 'react';
import { buildHighlightSegments, extractSearchTerms } from './searchHighlightRanges';
import styles from './HighlightedText.module.css';

interface HighlightedTextProps {
    text?: string;
    query?: string;
}

const HighlightedText = ({ text = '', query = '' }: HighlightedTextProps) => {
    const segments = useMemo(
        () => buildHighlightSegments(text, extractSearchTerms(query)),
        [text, query]
    );

    return (
        <>
            {segments.map((segment, segmentIndex) =>
                segment.isMatch ? (
                    <mark key={segmentIndex} className={styles.match}>
                        {segment.text}
                    </mark>
                ) : (
                    <span key={segmentIndex}>{segment.text}</span>
                )
            )}
        </>
    );
};

export default HighlightedText;
