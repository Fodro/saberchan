import { OIDC_REALM, OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, OIDC_REDIRECT_URI } from "$env/static/private";
import * as arctic from "arctic";

export const codeVerifier = arctic.generateCodeVerifier();
export const keycloak = new arctic.KeyCloak(OIDC_REALM, OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, OIDC_REDIRECT_URI);
