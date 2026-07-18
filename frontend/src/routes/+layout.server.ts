import { MAIN_BACKEND_URL } from '$env/static/private';
import { beginLogin, redirectIfNeedsRefresh } from '$lib/server/auth';
import { loadTranslations, translations } from '$lib/translations';
import type { Board } from '$lib/types/board';
import type { Metadata } from '$lib/types/metadata';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ fetch, cookies, depends }) => {
	const boardsRes = await fetch(`${MAIN_BACKEND_URL}/api/v1/board`);
	const boards: Board[] = await boardsRes.json();

	const fingerprint = cookies.get('fingerprint');

	if (!fingerprint) {
		cookies.set('fingerprint', crypto.randomUUID(), { path: '/', httpOnly: true });
	}

	const recentBoards: string[] = cookies.get('recent-boards')?.split(',') ?? [];
	const meta: Metadata = {
		recentBoards,
	};

	depends('board:all');

	const locale = cookies.get('locale') || 'en';
	await loadTranslations(locale, '/');

	const { loginUrl, logoutUrl } = beginLogin(cookies);
	const session = redirectIfNeedsRefresh(cookies);

	if (!session.signed) {
		return {
			boards,
			meta,
			i18n: { locale, route: '/' },
			translations: translations.get(),
			loginUrl,
			logoutUrl,
			signed: false,
		};
	}

	return {
		boards,
		meta,
		i18n: { locale, route: '/' },
		translations: translations.get(),
		loginUrl,
		logoutUrl,
		signed: true,
		session: session.session,
		username: session.username,
		idToken: session.idToken,
	};
};
