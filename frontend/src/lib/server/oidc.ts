import { env } from '$env/dynamic/private';
import { createRemoteJWKSet, jwtVerify } from 'jose';

const OIDC_CLIENT_ID = env.OIDC_CLIENT_ID;
const OIDC_REALM = env.OIDC_REALM;
const OIDC_REALM_INTERNAL = env.OIDC_REALM_INTERNAL;

function trimSlash(url: string): string {
	return url.replace(/\/$/, '');
}

const publicIssuer = trimSlash(OIDC_REALM);
const internalIssuer = trimSlash(OIDC_REALM_INTERNAL || OIDC_REALM);

/** Fetch JWKS from the issuer the Node process can reach (Docker-internal Keycloak). */
const jwks = createRemoteJWKSet(
	new URL(`${internalIssuer}/protocol/openid-connect/certs`),
);

/**
 * Cryptographically verify a Keycloak access token.
 * Accepts either public or internal issuer (Keycloak may stamp either depending on
 * which hostname was used for the token exchange).
 */
export async function verifyAccessToken(token: string): Promise<boolean> {
	try {
		const { payload } = await jwtVerify(token, jwks, {
			issuer: [publicIssuer, internalIssuer],
		});
		// Keycloak access tokens use azp for the confidential client id.
		if (typeof payload.azp === 'string' && payload.azp !== OIDC_CLIENT_ID) {
			return false;
		}
		return true;
	} catch {
		return false;
	}
}
