import axios from 'axios';

export const isNetworkError = (error: unknown): boolean =>
    axios.isAxiosError(error) && error.response === undefined;

export const isClientError = (error: unknown): boolean => {
    if (!axios.isAxiosError(error) || error.response === undefined) return false;
    return error.response.status >= 400 && error.response.status < 500;
};
