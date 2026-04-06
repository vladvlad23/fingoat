import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { api } from '$lib/api';

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

		try {
			const { token } = await api.auth.register(fetch, email, password);
			cookies.set('token', token, { path: '/', httpOnly: true, sameSite: 'lax', maxAge: 15 * 60 });
		} catch (e) {
			return fail(409, { error: (e as Error).message });
		}

		redirect(302, '/');
	}
};
