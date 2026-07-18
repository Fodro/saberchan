import { backendUrl } from '$lib/server/backend';

export const GET = async () => {
	return fetch(backendUrl('/api/v1/captcha'));
};
