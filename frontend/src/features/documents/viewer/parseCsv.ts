const byteOrderMark = new RegExp(`^${String.fromCharCode(0xfeff)}`);
const delimiterCandidates = [',', ';', '\t'];

const detectDelimiter = (firstLine: string): string => {
    let bestDelimiter = ',';
    let bestCount = 0;
    for (const delimiter of delimiterCandidates) {
        const occurrences = firstLine.split(delimiter).length - 1;
        if (occurrences > bestCount) {
            bestDelimiter = delimiter;
            bestCount = occurrences;
        }
    }
    return bestDelimiter;
};

export const parseCsv = (source: string): string[][] => {
    const normalizedSource = source.replace(byteOrderMark, '').replace(/\r\n?/g, '\n');
    const delimiter = detectDelimiter(normalizedSource.split('\n', 1)[0] ?? '');
    const rows: string[][] = [];
    let currentRow: string[] = [];
    let currentCell = '';
    let isInsideQuotes = false;

    const endCell = () => {
        currentRow.push(currentCell);
        currentCell = '';
    };
    const endRow = () => {
        endCell();
        rows.push(currentRow);
        currentRow = [];
    };

    for (let position = 0; position < normalizedSource.length; position += 1) {
        const character = normalizedSource[position];
        if (isInsideQuotes) {
            if (character !== '"') {
                currentCell += character;
            } else if (normalizedSource[position + 1] === '"') {
                currentCell += '"';
                position += 1;
            } else {
                isInsideQuotes = false;
            }
            continue;
        }
        if (character === '"') isInsideQuotes = true;
        else if (character === delimiter) endCell();
        else if (character === '\n') endRow();
        else currentCell += character;
    }

    if (currentCell !== '' || currentRow.length > 0) endRow();
    return rows.filter((row) => row.length > 1 || row[0] !== '');
};
