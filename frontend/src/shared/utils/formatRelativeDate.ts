const MINUTE_IN_MS = 60_000;
const HOUR_IN_MS = 60 * MINUTE_IN_MS;
const DAY_IN_MS = 24 * HOUR_IN_MS;
const MAX_RELATIVE_DAYS = 7;

const resolveDocumentLanguageTag = (): string | undefined => {
    try {
        return Intl.getCanonicalLocales(document.documentElement.lang)[0];
    } catch {
        return undefined;
    }
};

/**
 * Formata um instante ISO 8601 no idioma do documento: relativo ("há 3 horas") até 7 dias,
 * data absoluta depois (`00/RN-44`). Devolve string vazia para instante inválido ou ausente.
 */
export const formatRelativeDate = (
    isoTimestamp: string | undefined,
    now: Date = new Date()
): string => {
    if (!isoTimestamp) return '';
    const moment = new Date(isoTimestamp);
    if (Number.isNaN(moment.getTime())) return '';

    const languageTag = resolveDocumentLanguageTag();
    const elapsedMs = Math.max(0, now.getTime() - moment.getTime());
    const relativeFormat = new Intl.RelativeTimeFormat(languageTag, { numeric: 'auto' });

    if (elapsedMs < MINUTE_IN_MS) return relativeFormat.format(0, 'second');
    if (elapsedMs < HOUR_IN_MS)
        return relativeFormat.format(-Math.floor(elapsedMs / MINUTE_IN_MS), 'minute');
    if (elapsedMs < DAY_IN_MS)
        return relativeFormat.format(-Math.floor(elapsedMs / HOUR_IN_MS), 'hour');
    if (elapsedMs <= MAX_RELATIVE_DAYS * DAY_IN_MS)
        return relativeFormat.format(-Math.floor(elapsedMs / DAY_IN_MS), 'day');
    return new Intl.DateTimeFormat(languageTag, { dateStyle: 'medium' }).format(moment);
};
