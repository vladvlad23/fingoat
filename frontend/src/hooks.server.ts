import type { Handle } from '@sveltejs/kit';
import { env } from '$env/dynamic/public';
import { api, extractCookieValue, tokenSecondsRemaining } from '$lib/api';

/** Decode the payload of a JWT without verifying the signature. */
function decodeJwtPayload(token: string): Record<string, unknown> | null {
	try {
		const parts = token.split('.');
		if (parts.length !== 3) return null;
		// Base64url → Base64 → JSON
		const base64 = parts[1].replace(/-/g, '+').replace(/_/g, '/');
		const json = atob(base64);
		return JSON.parse(json) as Record<string, unknown>;
	} catch {
		return null;
	}
}

/**
 * Returns true when the JWT's exp claim is in the past (or missing/malformed).
 * Uses a 10-second clock-skew buffer to refresh slightly before hard expiry.
 */
function isTokenExpired(token: string): boolean {
	const payload = decodeJwtPayload(token);
	if (!payload) return true;
	const exp = payload['exp'];
	if (typeof exp !== 'number') return true;
	return Date.now() / 1000 > exp - 10;
}

/**
 * Call POST /api/auth/refresh, forwarding the incoming Cookie header so the
 * httpOnly refresh_token cookie is sent automatically. On success, sets the
 * rotated refresh_token cookie on the response and returns the new access token.
 */
async function tryRefresh(
	event: Parameters<Handle>[0]['event']
): Promise<string | null> {
	const resp = await fetch(`${env.PUBLIC_API_URL}/api/auth/refresh`, {
		method: 'POST',
		headers: {
			// Forward all cookies so the httpOnly refresh_token reaches the backend.
			cookie: event.request.headers.get('cookie') ?? ''
		}
	});

	if (!resp.ok) return null;

	// Rotate the refresh_token cookie: forward what the backend sent back.
	const setCookieHeader = resp.headers.get('set-cookie');
	if (setCookieHeader) {
		const value = extractCookieValue(setCookieHeader, 'refresh_token');
		if (value) {
			event.cookies.set('refresh_token', value, {
				path: '/api/auth',
				httpOnly: true,
				sameSite: 'lax',
				maxAge: 30 * 24 * 60 * 60
			});
		}
	}

	const data = (await resp.json()) as { token?: string };
	return data.token ?? null;
}


export const handle: Handle = async ({ event, resolve }) => {
	event.locals.token = null;
	event.locals.user = null;

	let token = event.cookies.get('token') ?? null;

	if (token && isTokenExpired(token)) {
		// Proactively refresh before the route runs — only if a refresh cookie exists.
		if (event.cookies.get('refresh_token')) {
			const newToken = await tryRefresh(event);
			if (newToken) {
				event.cookies.set('token', newToken, {
					path: '/',
					httpOnly: true,
					sameSite: 'lax',
					maxAge: tokenSecondsRemaining(newToken)
				});
				token = newToken;
			} else {
				// Refresh token expired or revoked — drop the stale access token.
				event.cookies.delete('token', { path: '/' });
				token = null;
			}
		} else {
			// No refresh cookie — nothing to refresh with.
			event.cookies.delete('token', { path: '/' });
			token = null;
		}
	}

	if (token) {
		event.locals.token = token;
		try {
			const user = await api.user.me(event.fetch, token);
			event.locals.user = {
				id: user.id,
				email: user.email,
				monthlyIncome: user.monthlyIncome,
				paymentDay: user.paymentDay,
				currency: user.currency
			};
		} catch {
			// Token invalid on the backend (shouldn't happen after refresh, but be safe).
			event.cookies.delete('token', { path: '/' });
			event.locals.token = null;
		}
	}

	return resolve(event);
};
