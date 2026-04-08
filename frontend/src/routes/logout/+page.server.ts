import { redirect } from '@sveltejs/kit';
import type { Actions } from './$types';
import { env } from '$env/dynamic/public';

export const actions: Actions = {
	default: async ({ cookies, request }) => {
		// Tell the backend to revoke the refresh token. Forward cookies so the
		// httpOnly refresh_token cookie is included in the request.
		try {
			await fetch(`${env.PUBLIC_API_URL}/api/auth/logout`, {
				method: 'POST',
				headers: {
					cookie: request.headers.get('cookie') ?? ''
				}
			});
		} catch {
			// Backend unreachable — still clear local cookies and redirect.
		}

		cookies.delete('token', { path: '/' });
		// Clear the refresh token cookie (same attributes it was set with).
		cookies.delete('refresh_token', { path: '/api/auth' });
		redirect(302, '/login');
	}
};
