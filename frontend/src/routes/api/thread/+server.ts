import { MAIN_BACKEND_URL } from "$env/dynamic/private";
import type { Thread } from "$lib/types/thread";
import { error } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";
// import { verifyExp } from "$lib/helpers";
// import { jwtDecode } from "jwt-decode";

export const POST: RequestHandler = async ({request, cookies, fetch}) => {
	const body: Thread = await request.json();

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

	body.original_post.browser_fingerprint = fingerprint || '';
	body.original_post.ip = '0.0.0.0';
	body.original_post.sage = false;
	body.original_post.op_marker = true;

	// const token = cookies.get("accessToken");
	
	// if (token && !verifyExp(jwtDecode(token).exp)) {
	// 	body.is_admin = true;
	// }

	const res = await fetch(`${MAIN_BACKEND_URL}/api/v1/thread`, {
		method: 'POST',
		body: JSON.stringify(body),
	})

	return res;
}; 