import { cookieSecure, keycloakAuthorize, oidcRealmPublic } from '$lib/auth';
import { verifyExp } from '$lib/helpers';
import {
	AUTH_HOST,
	OIDC_CLIENT_ID,
} from '$env/static/private';
import * as arctic from 'arctic';
import { jwtDecode, type JwtPayload } from 'jwt-decode';
import type { Cookies } from '@sveltejs/kit';
import { redirect } from '@sveltejs/kit';

export const PKCE_VERIFIER_COOKIE = 'oidc_code_verifier';

export type AuthLayoutResult = {
	loginUrl: string;
	logoutUrl: string;
	signed: boolean;
	session?: string;
	username?: string;
	idToken?: string;
};

/** Create authorize URL and persist a per-login PKCE verifier in an httpOnly cookie. */
export function beginLogin(cookies: Cookies): { loginUrl: string; logoutUrl: string } {
	const state = arctic.generateState();
	let codeVerifier = cookies.get(PKCE_VERIFIER_COOKIE);
	if (!codeVerifier) {
		codeVerifier = arctic.generateCodeVerifier();
		cookies.set(PKCE_VERIFIER_COOKIE, codeVerifier, {
			path: '/',
			httpOnly: true,
			secure: cookieSecure,
			sameSite: 'lax',
			maxAge: 60 * 10,
		});
	}
	const scopes = ['openid', 'profile'];
	const url = keycloakAuthorize.createAuthorizationURL(state, codeVerifier, scopes);

	const logoutUrl = `${oidcRealmPublic}/protocol/openid-connect/logout?post_logout_redirect_uri=${AUTH_HOST}/admin/auth/signOut&client_id=${OIDC_CLIENT_ID}`;

	return { loginUrl: url.href, logoutUrl };
}

export function readSession(cookies: Cookies): {
	signed: boolean;
	needsRefresh: boolean;
	session?: string;
	username?: string;
	idToken?: string;
} {
	const token = cookies.get('accessToken');
	const logoutNeededRefresh = !token || verifyExp(jwtDecode(token).exp);

	if (logoutNeededRefresh) {
		const refreshToken = cookies.get('refreshToken');
		return { signed: false, needsRefresh: Boolean(refreshToken) };
	}

	const decodedJwt = jwtDecode(token) as JwtPayload & { name: string | undefined };
	return {
		signed: true,
		needsRefresh: false,
		session: decodedJwt.sub,
		username: decodedJwt.name,
		idToken: cookies.get('idToken'),
	};
}

export function redirectIfNeedsRefresh(cookies: Cookies) {
	const session = readSession(cookies);
	if (!session.signed && session.needsRefresh) {
		redirect(302, '/admin/auth/refresh');
	}
	return session;
}
