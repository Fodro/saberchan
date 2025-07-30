import { MAIN_BACKEND_URL } from "$env/static/private";
import type { Post } from "$lib/types/post";
import { error } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";

export const POST: RequestHandler = async ({request, cookies, fetch}) => {
	const body: Post = await request.json();

	const captchaRes = await fetch(`${MAIN_BACKEND_URL}/api/v1/captcha`, {
		method: 'POST',
		body: JSON.stringify(body.captcha)
	})

	const captchaJson = await captchaRes.json();

	if (!captchaJson.passed) {
		error(403, {
			"message": "Captcha failed"
		})
	}

	const fingerprint = cookies.get('fingerprint');

	body.browser_fingerprint = fingerprint || '';
	body.ip = '0.0.0.0';

	const res = await fetch(`${MAIN_BACKEND_URL}/api/v1/post/${body.thread_id}`, {
		method: 'POST',
		body: JSON.stringify(body),
	})

	return res;
}; 