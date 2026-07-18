import type { RequestHandler } from './$types';
import {
	assertBodySize,
	assertCaptcha,
	assertMultipartFiles,
	assertText,
	fingerprintFromCookies,
	proxyBackend,
} from '$lib/server/backend';

export const POST: RequestHandler = async ({ request, cookies, fetch }) => {
	assertBodySize(request);

	const form = await request.formData();
	const captcha = {
		token: String(form.get('captcha_token') ?? ''),
		input: String(form.get('captcha_input') ?? ''),
	};
	assertCaptcha(captcha);
	assertText(String(form.get('text') ?? ''));

	const files = form.getAll('files').filter((f): f is File => f instanceof File && f.size > 0);
	assertMultipartFiles(files);

	// Captcha is consumed once on the Go API — do not pre-validate here.
	const threadId = String(form.get('thread_id') ?? '');
	const out = new FormData();
	out.set('text', String(form.get('text') ?? ''));
	out.set('sage', String(form.get('sage') ?? 'false'));
	out.set('op_marker', String(form.get('op_marker') ?? 'false'));
	out.set('browser_fingerprint', fingerprintFromCookies(cookies));
	out.set('captcha_token', captcha.token);
	out.set('captcha_input', captcha.input);
	for (const file of files) {
		out.append('files', file, file.name);
	}

	return proxyBackend(fetch, `/api/v1/post/${threadId}`, {
		method: 'POST',
		body: out,
	});
};
