import { MAIN_BACKEND_URL } from '$env/static/private';
import type { Thread } from '$lib/types/thread';
import { error } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { base64ToArrayBuffer, verifyExp } from '$lib/helpers';
import { jwtDecode } from 'jwt-decode';
import {
	MAX_FILE_BYTES,
	MAX_FILES,
	MAX_JSON_BODY_CHARS,
	MAX_TEXT_CHARS,
	MAX_TITLE_CHARS,
} from '$lib/limits';

export const POST: RequestHandler = async ({ request, cookies, fetch }) => {
	const contentLength = Number(request.headers.get('content-length') || 0);
	if (contentLength > MAX_JSON_BODY_CHARS) {
		error(413, { message: 'Post is too large (try fewer or smaller images)' });
	}

	const body: Thread = await request.json();

	if (!body.captcha?.token?.trim() || !body.captcha?.input?.trim()) {
		error(400, { message: 'Solve the captcha before posting' });
	}

	const title = body.title?.trim() ?? '';
	const text = body.original_post?.text?.trim() ?? '';
	if (!title) error(400, { message: 'Title is required' });
	if (title.length > MAX_TITLE_CHARS) {
		error(400, { message: `Title is too long (max ${MAX_TITLE_CHARS} characters)` });
	}
	if (!text) error(400, { message: 'Post text is required' });
	if (text.length > MAX_TEXT_CHARS) {
		error(400, { message: `Post text is too long (max ${MAX_TEXT_CHARS} characters)` });
	}

	const attachments = body.original_post?.attachments ?? [];
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

	if (!body.original_post) {
		error(400, { message: 'original_post is required' });
	}

	const fingerprint = cookies.get('fingerprint');

	body.original_post.browser_fingerprint = fingerprint || '';
	body.original_post.ip = '0.0.0.0';
	body.original_post.sage = false;
	body.original_post.op_marker = true;

	const token = cookies.get('accessToken');

	if (token && !verifyExp(jwtDecode(token).exp)) {
		body.is_admin = true;
	}

	const res = await fetch(`${MAIN_BACKEND_URL}/api/v1/thread`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body),
	});

	if (!res.ok) {
		const msg = await res.text();
		error(res.status, { message: msg || 'Failed to create thread' });
	}

	return res;
};
