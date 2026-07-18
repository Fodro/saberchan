import { MAIN_BACKEND_URL } from '$env/static/private';
import { beginLogin, redirectIfNeedsRefresh } from '$lib/server/auth';
import { adminBackendHeaders, isAdminSession } from '$lib/server/backend';
import { loadFollowedSummary } from '$lib/server/followed';
import { loadTranslations, translations } from '$lib/translations';
import type { Board } from '$lib/types/board';
import type { Metadata } from '$lib/types/metadata';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ fetch, cookies, depends }) => {
	const boardsRes = await fetch(`${MAIN_BACKEND_URL}/api/v1/board`, {
		headers: await adminBackendHeaders(cookies),
	});
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
	depends('followed:list');

	const followed = await loadFollowedSummary(fetch, cookies);

	const locale = cookies.get('locale') || 'en';
	await loadTranslations(locale, '/');

	const { loginUrl, logoutUrl } = beginLogin(cookies);
	const session = redirectIfNeedsRefresh(cookies);
	const isAdmin = await isAdminSession(cookies);

	if (!session.signed) {
		return {
			boards,
			meta,
			followed,
			i18n: { locale, route: '/' },
			translations: translations.get(),
			loginUrl,
			logoutUrl,
			signed: false,
			isAdmin: false,
		};
	}

	return {
		boards,
		meta,
		followed,
		i18n: { locale, route: '/' },
		translations: translations.get(),
		loginUrl,
		logoutUrl,
		signed: true,
		isAdmin,
		session: session.session,
		username: session.username,
		idToken: session.idToken,
	};
};
