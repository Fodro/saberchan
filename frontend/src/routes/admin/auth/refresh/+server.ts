import { keycloak } from "$lib/auth";
import { redirect } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";

export const GET: RequestHandler = async ({cookies}) => {
	const refreshTokenOld = cookies.get("refreshToken");
	if (refreshTokenOld) {
		let tokens;
		try {
			tokens = await keycloak.refreshAccessToken(refreshTokenOld);
		} catch(e) {
			console.log(e);
			redirect(302, '/admin/auth/signOut');
		}
		const accessToken = tokens.accessToken();
		const accessTokenExpiresAt = tokens.accessTokenExpiresAt();
		const refreshToken = tokens.refreshToken();

		cookies.set("accessToken", accessToken, {
			path: '/admin',
			httpOnly: true,
			maxAge: accessTokenExpiresAt.getTime() - Date.now(),
			secure: true,
			sameSite: 'strict',
			expires: accessTokenExpiresAt,
		})

		if ("refresh_expires_in" in tokens.data && typeof tokens.data.refresh_expires_in === "number") {
			const refreshTokenExpiresIn = new Date(tokens.data.refresh_expires_in * 1000);
			cookies.set("refreshToken", refreshToken, {
				path: '/admin',
				httpOnly: true,
				secure: true,
				sameSite: 'strict',
				maxAge: refreshTokenExpiresIn.getTime() - Date.now(),
				expires: refreshTokenExpiresIn,
			})
		} else {
			cookies.set("refreshToken", refreshToken, {
				path: '/admin',
				httpOnly: true,
				secure: true,
				sameSite: 'strict',
			})
		}
		redirect(302, '/admin');
	} else {
		redirect(302, '/admin/auth/signOut');
	}

}; 