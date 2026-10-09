export type HighlightSegment = { text: string; isMatch: boolean };

const combiningMarksPattern = /[̀-ͯ]/g;

const foldCharacter = (character: string) =>
    character.normalize('NFD').replace(combiningMarksPattern, '').toLowerCase();

export const extractSearchTerms = (query: string): string[] =>
    foldCharacter(query).split(/\s+/).filter(Boolean);

export const buildHighlightSegments = (text: string, terms: string[]): HighlightSegment[] => {
    if (!text) {
        return [];
    }

    if (terms.length === 0) {
        return [{ text, isMatch: false }];
    }

    const originalCharacters = Array.from(text);
    let foldedText = '';
    const foldedOffsetToCharacterIndex: number[] = [];
    originalCharacters.forEach((character, characterIndex) => {
        const folded = foldCharacter(character);
        for (let offset = 0; offset < folded.length; offset += 1) {
            foldedOffsetToCharacterIndex.push(characterIndex);
        }
        foldedText += folded;
    });

    const isCharacterMatched: boolean[] = originalCharacters.map(() => false);
    terms.forEach((term) => {
        let searchFrom = 0;
        let matchStart = foldedText.indexOf(term, searchFrom);
        while (matchStart !== -1) {
            for (let offset = matchStart; offset < matchStart + term.length; offset += 1) {
                const characterIndex = foldedOffsetToCharacterIndex[offset];
                if (characterIndex !== undefined) {
                    isCharacterMatched[characterIndex] = true;
                }
            }
            searchFrom = matchStart + term.length;
            matchStart = foldedText.indexOf(term, searchFrom);
        }
    });

    const segments: HighlightSegment[] = [];
    originalCharacters.forEach((character, characterIndex) => {
        const isMatch = isCharacterMatched[characterIndex] === true;
        const lastSegment = segments[segments.length - 1];
        if (lastSegment && lastSegment.isMatch === isMatch) {
            lastSegment.text += character;
            return;
        }
        segments.push({ text: character, isMatch });
    });
    return segments;
};
