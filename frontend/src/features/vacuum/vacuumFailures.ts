import axios from 'axios';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';

const notConfiguredStatusCode = 409;

export const isVacuumNotConfiguredError = (error: unknown): boolean =>
    axios.isAxiosError(error) && error.response?.status === notConfiguredStatusCode;

export const resolveVacuumFailureMessage = (error: unknown, connectionMessage: string): string =>
    extractBackendErrorMessage(error) ?? connectionMessage;
