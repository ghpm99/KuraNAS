import { Box, Link, Typography } from '@mui/material';
import type { ReactNode } from 'react';
import { parseMarkdown, type MarkdownBlock, type MarkdownInline } from './parseMarkdown';

const maxRenderedHeadingLevel = 6;

const renderInlines = (inlines: MarkdownInline[]): ReactNode[] =>
    inlines.map((inline, position) => {
        if (inline.kind === 'text') return inline.text;
        if (inline.kind === 'code') {
            return (
                <Box component="code" key={position} sx={{ fontFamily: 'monospace' }}>
                    {inline.text}
                </Box>
            );
        }
        if (inline.kind === 'strong') {
            return <strong key={position}>{renderInlines(inline.children)}</strong>;
        }
        if (inline.kind === 'emphasis') {
            return <em key={position}>{renderInlines(inline.children)}</em>;
        }
        return (
            <Link key={position} href={inline.href} target="_blank" rel="noopener noreferrer">
                {inline.text}
            </Link>
        );
    });

const renderBlock = (block: MarkdownBlock, position: number): ReactNode => {
    if (block.kind === 'heading') {
        const headingLevel = Math.min(block.level + 1, maxRenderedHeadingLevel);
        return (
            <Typography
                key={position}
                component={`h${headingLevel}` as 'h2'}
                variant="h6"
                sx={{ mt: 2, mb: 1 }}
            >
                {renderInlines(block.children)}
            </Typography>
        );
    }
    if (block.kind === 'list') {
        return (
            <Box key={position} component={block.isOrdered ? 'ol' : 'ul'} sx={{ my: 1, pl: 3 }}>
                {block.items.map((item, itemPosition) => (
                    <li key={itemPosition}>{renderInlines(item)}</li>
                ))}
            </Box>
        );
    }
    if (block.kind === 'code') {
        return (
            <Box
                key={position}
                component="pre"
                tabIndex={0}
                sx={{
                    m: 0,
                    my: 1,
                    p: 1.5,
                    overflow: 'auto',
                    fontFamily: 'monospace',
                    fontSize: '0.85rem',
                    borderRadius: 1,
                    border: '1px solid',
                    borderColor: 'divider',
                }}
            >
                {block.text}
            </Box>
        );
    }
    return (
        <Typography key={position} variant="body1" sx={{ my: 1 }}>
            {renderInlines(block.children)}
        </Typography>
    );
};

const DocumentMarkdownView = ({ source }: { source: string }) => (
    <Box sx={{ minWidth: 0, overflowWrap: 'anywhere' }}>
        {parseMarkdown(source ?? '').map(renderBlock)}
    </Box>
);

export default DocumentMarkdownView;
