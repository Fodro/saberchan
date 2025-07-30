import { MAIN_BACKEND_URL } from '$env/dynamic/private';
import { loadTranslations, translations } from '$lib/translations';
import type { Board } from '$lib/types/board';
import type { Metadata } from '$lib/types/metadata';
import type { LayoutServerLoad } from './$types';
import { AUTH_HOST, OIDC_CLIENT_ID, OIDC_REALM } from "$env/dynamic/private";
// import { codeVerifier, keycloak } from "$lib/auth";
// import * as arctic from "arctic";
// import { redirect } from '@sveltejs/kit';
// import { jwtDecode, type JwtPayload } from 'jwt-decode';
// import { verifyExp } from '$lib/helpers';

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
	}

	depends('board:all');

	const locale = cookies.get('locale') || 'en';
	await loadTranslations(locale, '/');

	// const state = arctic.generateState();
	// const scopes = ["openid", "profile"];
	// const url = keycloak.createAuthorizationURL(state, codeVerifier, scopes);
	const logoutUrl = `${OIDC_REALM}/protocol/openid-connect/logout?post_logout_redirect_uri=${AUTH_HOST}/admin/auth/signOut&client_id=${OIDC_CLIENT_ID}`

	// const token = cookies.get("accessToken");

	// if (!token || verifyExp(jwtDecode(token).exp)) {
	// 	const refreshToken = cookies.get("refreshToken");
	// 	if (refreshToken) {
	// 		redirect(302, "/admin/auth/refresh");
	// 	}
	// 	return {
	// 		boards, meta,
	// 		i18n: { locale, route: '/' },
	// 		translations: translations.get(),
	// 		loginUrl: "/",
	// 		logoutUrl,
	// 		signed: false,
	// 	};
	// }
	// const decodedJwt = (jwtDecode(token) as JwtPayload & {name: string | undefined});
	// const idToken = cookies.get("idToken");

	return {
		boards, meta, 
		i18n: { locale, route: '/' },
		translations: translations.get(),
		loginUrl: "/",
		logoutUrl,
		signed: false,
		// session: decodedJwt.sub,
		// username: decodedJwt.name,
		// idToken,
	 };
};