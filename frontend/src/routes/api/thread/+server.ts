import type { Thread } from '$lib/types/thread';
import { error } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import {
	assertAttachments,
	assertBodySize,
	assertCaptcha,
	assertText,
	assertTitle,
	fingerprintFromCookies,
	isAdminSession,
	proxyBackend,
	validateCaptchaWithBackend,
} from '$lib/server/backend';

export const POST: RequestHandler = async ({ request, cookies, fetch }) => {
	assertBodySize(request);

	const body: Thread = await request.json();
	assertCaptcha(body.captcha);
	assertTitle(body.title);

	if (!body.original_post) {
		error(400, { message: 'original_post is required' });
	}

	assertText(body.original_post.text);
	assertAttachments(body.original_post.attachments ?? []);

	await validateCaptchaWithBackend(fetch, {
		token: body.captcha!.token!,
		input: body.captcha!.input!,
	});

	body.original_post.browser_fingerprint = fingerprintFromCookies(cookies);
	delete (body.original_post as { ip?: string }).ip;
	body.original_post.sage = false;
	body.original_post.op_marker = true;
	body.is_admin = isAdminSession(cookies);

	return proxyBackend(fetch, '/api/v1/thread', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body),
	});
};
