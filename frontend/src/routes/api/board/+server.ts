import { MAIN_BACKEND_URL } from "$env/dynamic/private";
// import { verifyExp } from "$lib/helpers";
import type { Board } from "$lib/types/board";
// import { jwtDecode } from "jwt-decode";
import type { RequestHandler } from "./$types";
// import { error } from "@sveltejs/kit";

export const POST: RequestHandler = async ({ request, fetch}) => {
	// const token = cookies.get("accessToken");

	// if (!token || verifyExp(jwtDecode(token).exp)) {
	// 	error(401, {
	// 		message: "Unauthorized",
	// 	})
	// }
	const body: Board = await request.json();

	body.alias = body.alias.replace("/", "")

	const res = await fetch(`${MAIN_BACKEND_URL}/api/v1/board`, {
		method: 'POST',
		body: JSON.stringify(body),
	})

	return res;
}; 