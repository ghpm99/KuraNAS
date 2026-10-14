import { useCallback, useState } from 'react';

const failedPosterUrls = new Set<string>();

export const resetFailedPosterUrls = () => failedPosterUrls.clear();

type LocalPoster = {
    imageUrl: string;
    isPosterShown: boolean;
    handlePosterError: () => void;
};

export const useLocalPoster = (posterUrl: string, fallbackUrl: string): LocalPoster => {
    const [failedUrl, setFailedUrl] = useState<string | null>(null);
    const hasPosterFailed = failedUrl === posterUrl || failedPosterUrls.has(posterUrl);

    const handlePosterError = useCallback(() => {
        failedPosterUrls.add(posterUrl);
        setFailedUrl(posterUrl);
    }, [posterUrl]);

    return {
        imageUrl: hasPosterFailed ? fallbackUrl : posterUrl,
        isPosterShown: !hasPosterFailed,
        handlePosterError,
    };
};
