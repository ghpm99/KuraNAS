export type MarkdownInline =
    | { kind: 'text'; text: string }
    | { kind: 'strong'; children: MarkdownInline[] }
    | { kind: 'emphasis'; children: MarkdownInline[] }
    | { kind: 'code'; text: string }
    | { kind: 'link'; href: string; text: string };

export type MarkdownBlock =
    | { kind: 'heading'; level: number; children: MarkdownInline[] }
    | { kind: 'paragraph'; children: MarkdownInline[] }
    | { kind: 'list'; isOrdered: boolean; items: MarkdownInline[][] }
    | { kind: 'code'; text: string };

const inlinePattern = /`([^`]+)`|\*\*([^*]+)\*\*|\*([^*\s][^*]*)\*|\[([^\]]+)\]\(([^)\s]+)\)/g;
const safeLinkPattern = /^https?:\/\/[^\s]+$/i;
const headingPattern = /^(#{1,6})\s+(.*)$/;
const unorderedItemPattern = /^\s*[-*+]\s+(.*)$/;
const orderedItemPattern = /^\s*\d+[.)]\s+(.*)$/;
const fencePattern = /^\s*```/;

export const isSafeLinkHref = (href: string): boolean => safeLinkPattern.test(href);

export const parseMarkdownInlines = (source: string): MarkdownInline[] => {
    const inlines: MarkdownInline[] = [];
    let consumedUntil = 0;
    const pattern = new RegExp(inlinePattern.source, 'g');
    let match = pattern.exec(source);
    while (match !== null) {
        if (match.index > consumedUntil) {
            inlines.push({ kind: 'text', text: source.slice(consumedUntil, match.index) });
        }
        inlines.push(toInline(match));
        consumedUntil = match.index + match[0].length;
        match = pattern.exec(source);
    }
    if (consumedUntil < source.length) {
        inlines.push({ kind: 'text', text: source.slice(consumedUntil) });
    }
    return inlines;
};

const toInline = (match: RegExpExecArray): MarkdownInline => {
    const [, codeText, strongText, emphasisText, linkText, linkHref] = match;
    if (codeText !== undefined) return { kind: 'code', text: codeText };
    if (strongText !== undefined) {
        return { kind: 'strong', children: parseMarkdownInlines(strongText) };
    }
    if (emphasisText !== undefined) {
        return { kind: 'emphasis', children: parseMarkdownInlines(emphasisText) };
    }
    const safeLinkText = linkText ?? '';
    if (linkHref !== undefined && isSafeLinkHref(linkHref)) {
        return { kind: 'link', href: linkHref, text: safeLinkText };
    }
    return { kind: 'text', text: safeLinkText };
};

export const parseMarkdown = (source: string): MarkdownBlock[] => {
    const blocks: MarkdownBlock[] = [];
    const lines = source.replace(/\r\n?/g, '\n').split('\n');
    let paragraphLines: string[] = [];
    let currentList: Extract<MarkdownBlock, { kind: 'list' }> | null = null;
    let codeLines: string[] | null = null;

    const flushParagraph = () => {
        if (paragraphLines.length === 0) return;
        blocks.push({
            kind: 'paragraph',
            children: parseMarkdownInlines(paragraphLines.join(' ')),
        });
        paragraphLines = [];
    };
    const flushList = () => {
        if (currentList) blocks.push(currentList);
        currentList = null;
    };
    const addListItem = (isOrdered: boolean, itemText: string) => {
        flushParagraph();
        if (currentList && currentList.isOrdered !== isOrdered) flushList();
        currentList ??= { kind: 'list', isOrdered, items: [] };
        currentList.items.push(parseMarkdownInlines(itemText));
    };

    for (const line of lines) {
        if (codeLines !== null) {
            if (fencePattern.test(line)) {
                blocks.push({ kind: 'code', text: codeLines.join('\n') });
                codeLines = null;
                continue;
            }
            codeLines.push(line);
            continue;
        }
        if (fencePattern.test(line)) {
            flushParagraph();
            flushList();
            codeLines = [];
            continue;
        }
        if (line.trim() === '') {
            flushParagraph();
            flushList();
            continue;
        }
        const headingMatch = headingPattern.exec(line);
        if (headingMatch) {
            flushParagraph();
            flushList();
            blocks.push({
                kind: 'heading',
                level: (headingMatch[1] ?? '').length,
                children: parseMarkdownInlines((headingMatch[2] ?? '').trim()),
            });
            continue;
        }
        const unorderedMatch = unorderedItemPattern.exec(line);
        if (unorderedMatch) {
            addListItem(false, unorderedMatch[1] ?? '');
            continue;
        }
        const orderedMatch = orderedItemPattern.exec(line);
        if (orderedMatch) {
            addListItem(true, orderedMatch[1] ?? '');
            continue;
        }
        flushList();
        paragraphLines.push(line.trim());
    }

    if (codeLines !== null) blocks.push({ kind: 'code', text: codeLines.join('\n') });
    flushParagraph();
    flushList();
    return blocks;
};
