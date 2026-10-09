import axios from 'axios';
import { getApiV1BaseUrl } from './apiUrl';
import { attachPlayerClientId } from './playerClientId';

export const apiBase = axios.create({
    baseURL: getApiV1BaseUrl(),
    headers: {
        'Content-Type': 'application/json',
    },
});

apiBase.interceptors.request.use(attachPlayerClientId);
