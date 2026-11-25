import { useState } from 'react';

export const useInFlightIds = () => {
    const [inFlightIds, setInFlightIds] = useState<ReadonlySet<number>>(new Set());

    const runExclusively = async (id: number, operation: () => Promise<void>) => {
        setInFlightIds((current) => new Set(current).add(id));
        try {
            await operation();
        } finally {
            setInFlightIds((current) => {
                const remaining = new Set(current);
                remaining.delete(id);
                return remaining;
            });
        }
    };

    return { isInFlight: (id: number) => inFlightIds.has(id), runExclusively };
};
