import { isSafeLinkHref, parseMarkdown, parseMarkdownInlines } from './parseMarkdown';

describe('parseMarkdown', () => {
    it('parses headings, paragraphs and joins wrapped lines', () => {
        expect(parseMarkdown('# Titulo\n\nprimeira\nsegunda')).toEqual([
            { kind: 'heading', level: 1, children: [{ kind: 'text', text: 'Titulo' }] },
            { kind: 'paragraph', children: [{ kind: 'text', text: 'primeira segunda' }] },
        ]);
    });

    it('parses unordered and ordered lists and splits them when the kind changes', () => {
        const blocks = parseMarkdown('- a\n* b\n1. c\n2) d');

        expect(blocks).toEqual([
            {
                kind: 'list',
                isOrdered: false,
                items: [[{ kind: 'text', text: 'a' }], [{ kind: 'text', text: 'b' }]],
            },
            {
                kind: 'list',
                isOrdered: true,
                items: [[{ kind: 'text', text: 'c' }], [{ kind: 'text', text: 'd' }]],
            },
        ]);
    });

    it('keeps fenced code verbatim, including an unterminated fence', () => {
        expect(parseMarkdown('```\n# nao e titulo\n- nem lista\n```')).toEqual([
            { kind: 'code', text: '# nao e titulo\n- nem lista' },
        ]);
        expect(parseMarkdown('texto\n```\nabre')).toEqual([
            { kind: 'paragraph', children: [{ kind: 'text', text: 'texto' }] },
            { kind: 'code', text: 'abre' },
        ]);
    });

    it('closes a paragraph when a list starts and handles CRLF', () => {
        expect(parseMarkdown('intro\r\n- item\r\n\r\nfim')).toEqual([
            { kind: 'paragraph', children: [{ kind: 'text', text: 'intro' }] },
            { kind: 'list', isOrdered: false, items: [[{ kind: 'text', text: 'item' }]] },
            { kind: 'paragraph', children: [{ kind: 'text', text: 'fim' }] },
        ]);
    });

    it('returns no blocks for an empty source', () => {
        expect(parseMarkdown('')).toEqual([]);
    });
});

describe('parseMarkdownInlines', () => {
    it('parses strong, emphasis, code and http links', () => {
        expect(parseMarkdownInlines('a **b** *c* `d` [e](https://exemplo.com/x) f')).toEqual([
            { kind: 'text', text: 'a ' },
            { kind: 'strong', children: [{ kind: 'text', text: 'b' }] },
            { kind: 'text', text: ' ' },
            { kind: 'emphasis', children: [{ kind: 'text', text: 'c' }] },
            { kind: 'text', text: ' ' },
            { kind: 'code', text: 'd' },
            { kind: 'text', text: ' ' },
            { kind: 'link', href: 'https://exemplo.com/x', text: 'e' },
            { kind: 'text', text: ' f' },
        ]);
    });

    it('turns links with unsafe schemes into plain text', () => {
        expect(parseMarkdownInlines('[clique](javascript:alert(1))')).toEqual([
            { kind: 'text', text: 'clique' },
            { kind: 'text', text: ')' },
        ]);
        expect(parseMarkdownInlines('[x](data:text/html;base64,AAAA)')).toEqual([
            { kind: 'text', text: 'x' },
        ]);
        expect(parseMarkdownInlines('[x](JAVASCRIPT:alert)')).toEqual([
            { kind: 'text', text: 'x' },
        ]);
    });

    it('keeps raw html as plain text', () => {
        expect(parseMarkdownInlines('<script>alert(1)</script>')).toEqual([
            { kind: 'text', text: '<script>alert(1)</script>' },
        ]);
    });
});

describe('isSafeLinkHref', () => {
    it.each([
        ['http://a.com', true],
        ['HTTPS://a.com/p?q=1', true],
        ['javascript:alert(1)', false],
        ['//a.com', false],
        ['ftp://a.com', false],
        ['https://a b', false],
    ])('%s -> %s', (href, expected) => {
        expect(isSafeLinkHref(href)).toBe(expected);
    });
});
