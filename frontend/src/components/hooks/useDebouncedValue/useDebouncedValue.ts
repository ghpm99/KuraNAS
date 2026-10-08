import { useEffect, useState } from 'react';

const useDebouncedValue = <TValue>(value: TValue, delayMs: number): TValue => {
    const [debouncedValue, setDebouncedValue] = useState(value);

    useEffect(() => {
        const timerId = setTimeout(() => setDebouncedValue(value), delayMs);
        return () => clearTimeout(timerId);
    }, [value, delayMs]);

    return debouncedValue;
};

export default useDebouncedValue;
