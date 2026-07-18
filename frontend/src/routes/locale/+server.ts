import { cookieSecure } from '$lib/auth';
import { isAppLocale } from '$lib/locales';
import type { Locale } from '$lib/types/metadata';
import type { RequestHandler } from './$types';
import { error } from '@sveltejs/kit';

export const POST: RequestHandler = async ({ request, cookies }) => {
	const body: Locale = await request.json();
	const { locale } = body;
	if (!locale || !isAppLocale(locale)) {
		error(400, { message: 'unsupported locale' });
	}
	cookies.set('locale', locale, {
		path: '/',
		httpOnly: false,
		sameSite: 'lax',
		secure: cookieSecure,
		maxAge: 60 * 60 * 24 * 365,
	});

	return new Response('', { status: 200 });
};
