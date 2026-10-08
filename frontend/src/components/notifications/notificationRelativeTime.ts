const MINUTE_IN_MS = 60_000;
const HOUR_IN_MS = 60 * MINUTE_IN_MS;
const DAY_IN_MS = 24 * HOUR_IN_MS;
const MAX_RELATIVE_DAYS = 30;

const resolveLanguageTag = (): string | undefined => {
    try {
        return Intl.getCanonicalLocales(document.documentElement.lang)[0];
    } catch {
        return undefined;
    }
};

export const formatNotificationRelativeTime = (
    createdAt: string,
    now: Date = new Date()
): string => {
    const createdDate = new Date(createdAt);
    if (Number.isNaN(createdDate.getTime())) return '';

    const languageTag = resolveLanguageTag();
    const elapsedMs = Math.max(0, now.getTime() - createdDate.getTime());
    const relativeFormat = new Intl.RelativeTimeFormat(languageTag, {
        numeric: 'auto',
        style: 'narrow',
    });

    if (elapsedMs < MINUTE_IN_MS) return relativeFormat.format(0, 'second');
    if (elapsedMs < HOUR_IN_MS)
        return relativeFormat.format(-Math.floor(elapsedMs / MINUTE_IN_MS), 'minute');
    if (elapsedMs < DAY_IN_MS)
        return relativeFormat.format(-Math.floor(elapsedMs / HOUR_IN_MS), 'hour');
    if (elapsedMs < MAX_RELATIVE_DAYS * DAY_IN_MS) {
        return relativeFormat.format(-Math.floor(elapsedMs / DAY_IN_MS), 'day');
    }
    return createdDate.toLocaleDateString(languageTag);
};
