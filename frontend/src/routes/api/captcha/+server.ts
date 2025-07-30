import { MAIN_BACKEND_URL } from "$env/static/private";

export const GET = async () => {
	const res = await fetch(`${MAIN_BACKEND_URL}/api/v1/captcha`);
	
	return res
};