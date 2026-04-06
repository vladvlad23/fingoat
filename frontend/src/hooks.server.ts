import type { Handle } from '@sveltejs/kit';
import { api } from '$lib/api';

export const handle: Handle = async ({ event, resolve }) => {
	const token = event.cookies.get('token') ?? null;
	event.locals.token = token;
	event.locals.user = null;

	if (token) {
		try {
			const user = await api.user.me(event.fetch, token);
			event.locals.user = { id: user.id, email: user.email, monthlyIncome: user.monthlyIncome, paymentDay: user.paymentDay };
		} catch {
			// token invalid — clear it
			event.cookies.delete('token', { path: '/' });
			event.locals.token = null;
		}
	}

	return resolve(event);
};
