import type { Post } from '$lib/types/post';
import type { RequestHandler } from './$types';
import {
	assertAttachments,
	assertBodySize,
	assertCaptcha,
	assertText,
	fingerprintFromCookies,
	proxyBackend,
	validateCaptchaWithBackend,
} from '$lib/server/backend';

export const POST: RequestHandler = async ({ request, cookies, fetch }) => {
	assertBodySize(request);

	const body: Post = await request.json();
	assertCaptcha(body.captcha);
	assertText(body.text);
	assertAttachments(body.attachments ?? []);

	await validateCaptchaWithBackend(fetch, {
		token: body.captcha!.token!,
		input: body.captcha!.input!,
	});

	body.browser_fingerprint = fingerprintFromCookies(cookies);
	delete (body as { ip?: string }).ip;

	return proxyBackend(fetch, `/api/v1/post/${body.thread_id}`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body),
	});
};
