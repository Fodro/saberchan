import { MAIN_BACKEND_URL } from "$env/static/private";
import type { Thread } from "$lib/types/thread";
import type { RequestHandler } from "./$types";

export const POST: RequestHandler = async ({request, cookies, fetch}) => {
	const body: Thread = await request.json();
	const fingerprint = cookies.get('fingerprint');

	body.original_post.browser_fingerprint = fingerprint || '';
	body.original_post.ip = '0.0.0.0';
	body.original_post.sage = false;
	body.original_post.op_marker = true;

	const res = await fetch(`${MAIN_BACKEND_URL}/api/v1/thread`, {
		method: 'POST',
		body: JSON.stringify(body),
	})

	return res;
}; 