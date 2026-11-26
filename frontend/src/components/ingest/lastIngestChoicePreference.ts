export type LastIngestChoice = {
    targetRoot: string;
    subfolder: string;
    preset: string;
};

export const lastIngestChoiceStorageKey = 'kuranas.ingest.lastChoice';

const isLastIngestChoice = (parsedValue: unknown): parsedValue is LastIngestChoice => {
    if (typeof parsedValue !== 'object' || parsedValue === null) return false;
    const candidate = parsedValue as Record<string, unknown>;
    return (
        typeof candidate.targetRoot === 'string' &&
        typeof candidate.subfolder === 'string' &&
        typeof candidate.preset === 'string'
    );
};

export const loadLastIngestChoice = (): LastIngestChoice | null => {
    try {
        const storedChoice = window.localStorage.getItem(lastIngestChoiceStorageKey);
        if (storedChoice === null) return null;
        const parsedChoice: unknown = JSON.parse(storedChoice);
        return isLastIngestChoice(parsedChoice) ? parsedChoice : null;
    } catch {
        return null;
    }
};

export const saveLastIngestChoice = (choice: LastIngestChoice): void => {
    try {
        window.localStorage.setItem(lastIngestChoiceStorageKey, JSON.stringify(choice));
    } catch {
        return;
    }
};
