import { codeVerifier, keycloak } from "$lib/auth";
import { redirect } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";

export const GET: RequestHandler = async ({request, cookies}) => {
	const urlParams = new URLSearchParams(request.url);
	const code = urlParams.get("code") || '';
	const tokens = await keycloak.validateAuthorizationCode(code, codeVerifier);
	const accessToken = tokens.accessToken();
	const accessTokenExpiresAt = tokens.accessTokenExpiresAt();
	const refreshToken = tokens.refreshToken();

	cookies.set("accessToken", accessToken, {
		path: '/',
		httpOnly: true,
		maxAge: accessTokenExpiresAt.getTime() - Date.now(),
		secure: true,
		sameSite: 'strict',
		expires: accessTokenExpiresAt,
	})
	
	cookies.set("idToken", tokens.idToken(), {
		path: '/',
		httpOnly: true,
		secure: true,
		sameSite: 'strict',
	});

	if ("refresh_expires_in" in tokens.data && typeof tokens.data.refresh_expires_in === "number") {
		const refreshTokenExpiresIn = new Date (tokens.data.refresh_expires_in * 1000);
		cookies.set("refreshToken", refreshToken, {
			path: '/',
			httpOnly: true,
			secure: true,
			sameSite: 'strict',
			maxAge: refreshTokenExpiresIn.getTime() - Date.now(),
			expires: refreshTokenExpiresIn,
		})
	} else {
		cookies.set("refreshToken", refreshToken, {
			path: '/',
			httpOnly: true,
			secure: true,
			sameSite: 'strict',
		})
	}

	redirect(302, '/');
}; 