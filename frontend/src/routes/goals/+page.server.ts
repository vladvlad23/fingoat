import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { api } from '$lib/api';

export const load: PageServerLoad = async ({ locals, fetch }) => {
	if (!locals.token || !locals.user) redirect(302, '/login');

	const goals = await api.goals.list(fetch, locals.token);
	return { goals };
};

export const actions: Actions = {
	create: async ({ request, locals, fetch }) => {
		if (!locals.token) return fail(401, { createError: 'Unauthorized' });
		const form = await request.formData();
		const title = (form.get('title') as string)?.trim();
		const targetAmount = (form.get('targetAmount') as string)?.trim();
		const deadline = (form.get('deadline') as string) || undefined;
		const currency = (form.get('currency') as string)?.trim() || undefined;

		if (!title || !targetAmount) return fail(400, { createError: 'Title and target amount are required' });

		try {
			await api.goals.create(fetch, locals.token, { title, targetAmount, deadline: deadline || undefined, currency });
		} catch (e) {
			return fail(400, { createError: (e as Error).message });
		}
	},
	delete: async ({ request, locals, fetch }) => {
		if (!locals.token) return fail(401, { error: 'Unauthorized' });
		const form = await request.formData();
		const id = Number(form.get('id'));
		try {
			await api.goals.delete(fetch, locals.token, id);
		} catch (e) {
			return fail(400, { error: (e as Error).message });
		}
	}
};
