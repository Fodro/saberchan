import { env } from '$env/dynamic/private';
import * as arctic from 'arctic';

const AUTH_HOST = env.AUTH_HOST;
const OIDC_CLIENT_ID = env.OIDC_CLIENT_ID;
const OIDC_CLIENT_SECRET = env.OIDC_CLIENT_SECRET;
const OIDC_REALM = env.OIDC_REALM;
const OIDC_REALM_INTERNAL = env.OIDC_REALM_INTERNAL;

const redirectURI = `${AUTH_HOST ?? ''}/admin/auth/signIn`;

/** Browser-facing issuer (authorization + logout links). */
export const oidcRealmPublic = OIDC_REALM ?? '';

/**
 * Server-side issuer for token/refresh calls from the Node process.
 * In Docker this is the in-network Keycloak service; locally it can match OIDC_REALM.
 */
const oidcRealmInternal = OIDC_REALM_INTERNAL || OIDC_REALM || '';

/** Used only to build the authorize URL the browser opens. */
export const keycloakAuthorize = new arctic.KeyCloak(
	oidcRealmPublic,
	OIDC_CLIENT_ID ?? '',
	OIDC_CLIENT_SECRET ?? '',
	redirectURI,
);

/** Used for code exchange + refresh (must be reachable from the frontend container). */
export const keycloak = new arctic.KeyCloak(
	oidcRealmInternal,
	OIDC_CLIENT_ID ?? '',
	OIDC_CLIENT_SECRET ?? '',
	redirectURI,
);

/** Local HTTP cannot set Secure cookies — otherwise login appears to fail after callback. */
export const cookieSecure = AUTH_HOST?.startsWith('https://') ?? false;
