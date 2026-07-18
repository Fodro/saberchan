import { MAIN_BACKEND_URL } from '$env/static/private';
import type { Post } from '$lib/types/post';
import { error } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { base64ToArrayBuffer } from '$lib/helpers';
import {
	MAX_FILE_BYTES,
	MAX_FILES,
	MAX_JSON_BODY_CHARS,
	MAX_TEXT_CHARS,
} from '$lib/limits';

export const POST: RequestHandler = async ({ request, cookies, fetch }) => {
	const contentLength = Number(request.headers.get('content-length') || 0);
	if (contentLength > MAX_JSON_BODY_CHARS) {
		error(413, { message: 'Post is too large (try fewer or smaller images)' });
	}

	const body: Post = await request.json();

	if (!body.captcha?.token?.trim() || !body.captcha?.input?.trim()) {
		error(400, { message: 'Solve the captcha before posting' });
	}

	const text = body.text?.trim() ?? '';
	if (!text) error(400, { message: 'Post text is required' });
	if (text.length > MAX_TEXT_CHARS) {
		error(400, { message: `Post text is too long (max ${MAX_TEXT_CHARS} characters)` });
	}

	const attachments = body.attachments ?? [];
	if (attachments.length > MAX_FILES) {
		error(400, { message: `Maximum file count is ${MAX_FILES}` });
	}
	for (const attachment of attachments) {
		const buf = base64ToArrayBuffer(attachment.body);
		if (buf.byteLength > MAX_FILE_BYTES) {
			error(413, { message: 'Maximum file size is 2MB' });
		}
	}

	const captchaRes = await fetch(`${MAIN_BACKEND_URL}/api/v1/captcha`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body.captcha),
	});

	const captchaJson = await captchaRes.json();

	if (!captchaJson.passed) {
		error(403, {
			message: 'Captcha failed',
		});
	}

	const fingerprint = cookies.get('fingerprint');

	body.browser_fingerprint = fingerprint || '';
	body.ip = '0.0.0.0';

	const res = await fetch(`${MAIN_BACKEND_URL}/api/v1/post/${body.thread_id}`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body),
	});

	if (!res.ok) {
		const msg = await res.text();
		error(res.status, { message: msg || 'Failed to create post' });
	}

	return res;
};
