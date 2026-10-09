const zeroDatePrefix = '0001-';

type OptionalDateWrapper = { Value?: unknown; HasValue?: unknown };

const isOptionalDateWrapper = (value: unknown): value is OptionalDateWrapper =>
    typeof value === 'object' && value !== null && 'HasValue' in value;

const extractRawDate = (value: unknown): string => {
    if (typeof value === 'string') return value;
    if (
        isOptionalDateWrapper(value) &&
        value.HasValue === true &&
        typeof value.Value === 'string'
    ) {
        return value.Value;
    }
    return '';
};

export const readOptionalDate = (value: unknown): string | null => {
    const rawDate = extractRawDate(value);
    if (rawDate === '' || rawDate.startsWith(zeroDatePrefix)) return null;
    return rawDate;
};
