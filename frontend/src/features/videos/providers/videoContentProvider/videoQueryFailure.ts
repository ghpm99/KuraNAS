import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';

export type VideoQueryFailure = {
    message?: string;
    retry: () => void;
};

type FailableQuery = {
    isError: boolean;
    error: unknown;
    hasData: boolean;
    refetch: () => unknown;
};

export const toVideoQueryFailure = ({
    isError,
    error,
    hasData,
    refetch,
}: FailableQuery): VideoQueryFailure | null => {
    if (!isError || hasData) {
        return null;
    }
    return {
        message: extractBackendErrorMessage(error),
        retry: () => {
            void refetch();
        },
    };
};
