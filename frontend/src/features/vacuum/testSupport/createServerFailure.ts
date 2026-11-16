import { AxiosError, type AxiosResponse } from 'axios';

export const createServerFailure = (statusCode: number, serverMessage: string): AxiosError =>
    new AxiosError('request failed', AxiosError.ERR_BAD_RESPONSE, undefined, undefined, {
        status: statusCode,
        data: { error: serverMessage },
    } as AxiosResponse);

export const createNetworkFailure = (): AxiosError => new AxiosError('Network Error');
