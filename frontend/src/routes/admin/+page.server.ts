import { jwtDecode } from "jwt-decode";
import type { PageServerLoad } from "./$types";
import { redirect } from "@sveltejs/kit";

export const load: PageServerLoad = async ({cookies}) => {
	const token = cookies.get("accessToken");
	if (!token) {
		const refreshToken = cookies.get("refreshToken");
		if (refreshToken) {
			redirect(302, "/admin/auth/refresh");
		}
		return {
			signed: false,
		};
	}
	const decodedJwt = jwtDecode(token);
	const idToken = cookies.get("idToken");

	return {
		signed: true,
		session: decodedJwt.sub,
		idToken,
	};
}