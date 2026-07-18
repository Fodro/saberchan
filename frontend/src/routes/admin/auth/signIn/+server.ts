import { codeVerifier, cookieSecure, keycloak } from '$lib/auth';
import { redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ url, cookies }) => {
	const code = url.searchParams.get('code') || '';
	const tokens = await keycloak.validateAuthorizationCode(code, codeVerifier);
	const accessToken = tokens.accessToken();
	const accessTokenExpiresAt = tokens.accessTokenExpiresAt();
	const refreshToken = tokens.refreshToken();

	cookies.set('accessToken', accessToken, {
		path: '/',
		httpOnly: true,
		maxAge: Math.max(1, Math.floor((accessTokenExpiresAt.getTime() - Date.now()) / 1000)),
		secure: cookieSecure,
		sameSite: 'lax',
		expires: accessTokenExpiresAt,
	});

	cookies.set('idToken', tokens.idToken(), {
		path: '/',
		httpOnly: true,
		secure: cookieSecure,
		sameSite: 'lax',
	});

	if ('refresh_expires_in' in tokens.data && typeof tokens.data.refresh_expires_in === 'number') {
		const refreshTokenExpiresIn = new Date(Date.now() + tokens.data.refresh_expires_in * 1000);
		cookies.set('refreshToken', refreshToken, {
			path: '/',
			httpOnly: true,
			secure: cookieSecure,
			sameSite: 'lax',
			maxAge: tokens.data.refresh_expires_in,
			expires: refreshTokenExpiresIn,
		});
	} else {
		cookies.set('refreshToken', refreshToken, {
			path: '/',
			httpOnly: true,
			secure: cookieSecure,
			sameSite: 'lax',
		});
	}

	redirect(302, '/');
};
