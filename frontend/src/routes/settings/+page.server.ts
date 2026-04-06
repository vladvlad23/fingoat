import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { api } from '$lib/api';

export const load: PageServerLoad = ({ locals }) => {
	if (!locals.token || !locals.user) redirect(302, '/login');
	return { user: locals.user };
};

export const actions: Actions = {
	default: async ({ request, locals, fetch }) => {
		if (!locals.token) return fail(401, { error: 'Unauthorized' });
		const form = await request.formData();
		const monthlyIncome = (form.get('monthlyIncome') as string)?.trim();
		const paymentDayRaw = (form.get('paymentDay') as string)?.trim();
		const paymentDay = paymentDayRaw ? parseInt(paymentDayRaw, 10) : undefined;

		if (!monthlyIncome) return fail(400, { error: 'Monthly income is required' });

		try {
			const user = await api.user.updateSettings(fetch, locals.token, monthlyIncome, paymentDay);
			return { success: true, user };
		} catch (e) {
			return fail(400, { error: (e as Error).message });
		}
	}
};
