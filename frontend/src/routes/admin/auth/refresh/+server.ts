import { cookieSecure, keycloak } from '$lib/auth';
import { redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ cookies, request }) => {
	const refreshTokenOld = cookies.get('refreshToken');
	if (refreshTokenOld) {
		let tokens;
		try {
			tokens = await keycloak.refreshAccessToken(refreshTokenOld);
		} catch (e) {
			console.log(e);
			redirect(302, '/admin/auth/signOut');
		}
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
		// Redirect back to the referring page (or /admin if none)
		const referer = request.headers.get('referer');
		const redirectTo = referer ? new URL(referer).pathname : '/';
		redirect(302, redirectTo);
	} else {
		redirect(302, '/admin/auth/signOut');
	}
};
