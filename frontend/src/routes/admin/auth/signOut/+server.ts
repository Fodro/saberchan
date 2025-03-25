import { redirect } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";

export const GET: RequestHandler = async ({cookies}) => {
	cookies.delete("accessToken", {path: "/admin"});
	cookies.delete("refreshToken", { path: "/admin" });
	cookies.delete("idToken", { path: "/admin" });

	redirect(302, '/admin');
}; 