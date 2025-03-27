import { redirect } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";

export const GET: RequestHandler = async ({cookies}) => {
	cookies.delete("accessToken", {path: "/"});
	cookies.delete("refreshToken", { path: "/" });
	cookies.delete("idToken", { path: "/" });

	redirect(302, '/');
}; 