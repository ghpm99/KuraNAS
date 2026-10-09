const ellipsis = '…';

export const formatDocumentSnippet = (rawSnippet: string | undefined): string => {
    const snippetText = (rawSnippet ?? '').replace(/\s+/g, ' ').trim();
    if (snippetText === '') return '';
    return `${ellipsis}${snippetText}${ellipsis}`;
};
