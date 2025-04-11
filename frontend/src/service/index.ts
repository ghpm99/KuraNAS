import axios from 'axios';

export const apiFile = axios.create({
	baseURL: `${import.meta.env.VITE_API_URL}/api/v1/files`,
	headers: {
		'Content-Type': 'application/json',
	},
});
