const timePattern = /^(\d{1,2}):(\d{2})$/;

export const computeNextShutdownExecution = (configuredTime: string, now: Date): Date | null => {
    const match = timePattern.exec(configuredTime ?? '');
    if (!match) return null;
    const hours = Number(match[1]);
    const minutes = Number(match[2]);
    if (hours > 23 || minutes > 59) return null;

    const nextExecution = new Date(now);
    nextExecution.setHours(hours, minutes, 0, 0);
    if (nextExecution.getTime() <= now.getTime()) {
        nextExecution.setDate(nextExecution.getDate() + 1);
    }
    return nextExecution;
};
