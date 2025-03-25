import { OIDC_CLIENT_ID, OIDC_LOGOUT_REDIRECT_URI, OIDC_REALM } from "$env/static/private";
import { codeVerifier, keycloak } from "$lib/auth";
import type { LayoutServerLoad } from "./$types";
import * as arctic from "arctic";

export const load: LayoutServerLoad = async () => {
	const state = arctic.generateState();
	const scopes = ["openid", "profile"];
	const url = keycloak.createAuthorizationURL(state, codeVerifier, scopes);
	const logoutUrl = `${OIDC_REALM}/protocol/openid-connect/logout?redirect_uri=${OIDC_LOGOUT_REDIRECT_URI}&client_id=${OIDC_CLIENT_ID}`

	return {
		loginUrl: url.href,
		logoutUrl,
	}
};