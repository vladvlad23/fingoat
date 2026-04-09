import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { extractCookieValue, tokenSecondsRemaining, type ApiError } from '$lib/api';
import { env } from '$env/dynamic/public';

export const load: PageServerLoad = ({ locals }) => {
	if (locals.user) redirect(302, '/');
};

export const actions: Actions = {
	default: async ({ request, cookies, fetch }) => {
		const form = await request.formData();
		const email = form.get('email') as string;
		const password = form.get('password') as string;

		if (!email || !password) return fail(400, { error: 'Email and password are required' });
		if (password.length < 6) return fail(400, { error: 'Password must be at least 6 characters' });

		const res = await fetch(`${env.PUBLIC_API_URL}/api/auth/register`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ email, password })
		});

		const data = await res.json();
		if (!res.ok) {
			const message = (data as ApiError).error ?? 'Registration failed';
			const status = res.status === 403 ? 403 : 409;
			return fail(status, { error: message });
		}

		const { token } = data as { token: string };
		cookies.set('token', token, { path: '/', httpOnly: true, sameSite: 'lax', maxAge: tokenSecondsRemaining(token) });

		const setCookie = res.headers.get('set-cookie');
		if (setCookie) {
			const refreshToken = extractCookieValue(setCookie, 'refresh_token');
			if (refreshToken) {
				cookies.set('refresh_token', refreshToken, { path: '/api/auth', httpOnly: true, sameSite: 'lax', maxAge: 30 * 24 * 60 * 60 });
			}
		}

		redirect(302, '/');
	}
};
