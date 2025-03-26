import { AUTH_HOST, OIDC_CLIENT_ID, OIDC_REALM } from "$env/static/private";
import { codeVerifier, keycloak } from "$lib/auth";
import type { LayoutServerLoad } from "./$types";
import * as arctic from "arctic";

export const load: LayoutServerLoad = async () => {
	const state = arctic.generateState();
	const scopes = ["openid", "profile"];
	const url = keycloak.createAuthorizationURL(state, codeVerifier, scopes);
	const logoutUrl = `${OIDC_REALM}/protocol/openid-connect/logout?post_logout_redirect_uri=${AUTH_HOST}/admin/auth/signOut&client_id=${OIDC_CLIENT_ID}`

	return {
		loginUrl: url.href,
		logoutUrl,
	}
};