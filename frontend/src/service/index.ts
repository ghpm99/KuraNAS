import axios from 'axios';
import { getApiV1BaseUrl } from './apiUrl';

export const apiBase = axios.create({
	baseURL: getApiV1BaseUrl(),
	headers: {
		'Content-Type': 'application/json',
	},
});
